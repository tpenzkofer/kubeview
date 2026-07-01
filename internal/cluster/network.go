package cluster

import (
	"context"
	"fmt"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceInfo summarises a Service and the pods currently backing it.
type ServiceInfo struct {
	Namespace string
	Name      string
	Type      string
	ClusterIP string
	Ports     []string // "80/TCP→8080"
	Selector  string
	Endpoints []EndpointInfo
}

// EndpointInfo is one address backing a Service.
type EndpointInfo struct {
	IP      string
	PodName string
	Ready   bool
}

// IngressInfo summarises an Ingress.
type IngressInfo struct {
	Namespace string
	Name      string
	Class     string
	Rules     []string // "host/path → svc:port"
	Address   string
}

// NetPolInfo summarises a NetworkPolicy.
type NetPolInfo struct {
	Namespace   string
	Name        string
	PodSelector string
	Ingress     int
	Egress      int
}

// collectNetwork fills the networking sections of a Snapshot. Errors are
// swallowed per-resource so a missing API (e.g. no ingress controller) does not
// break the whole view.
func (c *Client) collectNetwork(ctx context.Context, namespace string, snap *Snapshot) {
	// endpoints first, keyed by ns/name (== service name)
	type epKey struct{ ns, name string }
	eps := map[epKey][]EndpointInfo{}
	if el, err := c.Clientset.CoreV1().Endpoints(namespace).List(ctx, metav1.ListOptions{}); err == nil {
		for _, e := range el.Items {
			k := epKey{e.Namespace, e.Name}
			for _, sub := range e.Subsets {
				for _, a := range sub.Addresses {
					eps[k] = append(eps[k], endpointFrom(a, true))
				}
				for _, a := range sub.NotReadyAddresses {
					eps[k] = append(eps[k], endpointFrom(a, false))
				}
			}
		}
	}

	if sl, err := c.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{}); err == nil {
		for _, s := range sl.Items {
			si := ServiceInfo{
				Namespace: s.Namespace,
				Name:      s.Name,
				Type:      string(s.Spec.Type),
				ClusterIP: s.Spec.ClusterIP,
				Selector:  selectorString(s.Spec.Selector),
				Endpoints: eps[epKey{s.Namespace, s.Name}],
			}
			for _, p := range s.Spec.Ports {
				port := fmt.Sprintf("%d/%s", p.Port, p.Protocol)
				if p.TargetPort.String() != "" && p.TargetPort.String() != fmt.Sprintf("%d", p.Port) {
					port += "→" + p.TargetPort.String()
				}
				if p.NodePort != 0 {
					port += fmt.Sprintf(" (node:%d)", p.NodePort)
				}
				si.Ports = append(si.Ports, port)
			}
			snap.Services = append(snap.Services, si)
		}
		sort.Slice(snap.Services, func(i, j int) bool {
			if snap.Services[i].Namespace != snap.Services[j].Namespace {
				return snap.Services[i].Namespace < snap.Services[j].Namespace
			}
			return snap.Services[i].Name < snap.Services[j].Name
		})
	}

	if il, err := c.Clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{}); err == nil {
		for _, ing := range il.Items {
			ii := IngressInfo{Namespace: ing.Namespace, Name: ing.Name}
			if ing.Spec.IngressClassName != nil {
				ii.Class = *ing.Spec.IngressClassName
			}
			for _, lb := range ing.Status.LoadBalancer.Ingress {
				if lb.IP != "" {
					ii.Address = lb.IP
				} else if lb.Hostname != "" {
					ii.Address = lb.Hostname
				}
			}
			for _, r := range ing.Spec.Rules {
				host := r.Host
				if host == "" {
					host = "*"
				}
				if r.HTTP != nil {
					for _, p := range r.HTTP.Paths {
						svc := ""
						if p.Backend.Service != nil {
							svc = p.Backend.Service.Name
							if p.Backend.Service.Port.Number != 0 {
								svc += fmt.Sprintf(":%d", p.Backend.Service.Port.Number)
							} else if p.Backend.Service.Port.Name != "" {
								svc += ":" + p.Backend.Service.Port.Name
							}
						}
						ii.Rules = append(ii.Rules, fmt.Sprintf("%s%s → %s", host, p.Path, svc))
					}
				}
			}
			snap.Ingresses = append(snap.Ingresses, ii)
		}
	}

	if nl, err := c.Clientset.NetworkingV1().NetworkPolicies(namespace).List(ctx, metav1.ListOptions{}); err == nil {
		for _, np := range nl.Items {
			snap.NetPols = append(snap.NetPols, NetPolInfo{
				Namespace:   np.Namespace,
				Name:        np.Name,
				PodSelector: selectorString(np.Spec.PodSelector.MatchLabels),
				Ingress:     len(np.Spec.Ingress),
				Egress:      len(np.Spec.Egress),
			})
		}
	}
}

func endpointFrom(a corev1.EndpointAddress, ready bool) EndpointInfo {
	ei := EndpointInfo{IP: a.IP, Ready: ready}
	if a.TargetRef != nil {
		ei.PodName = a.TargetRef.Name
	}
	return ei
}

func selectorString(m map[string]string) string {
	if len(m) == 0 {
		return "<none>"
	}
	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, k+"="+v)
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}
