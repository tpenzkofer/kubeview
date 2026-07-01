package cluster

import (
	"context"
	"sort"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EventInfo is one cluster event, flattened for display.
type EventInfo struct {
	Type      string // Normal / Warning
	Reason    string
	Namespace string
	Object    string // kind/name
	Message   string
	Count     int32
	Age       time.Duration
}

// collectEvents fills snap.Events with the most recent events (newest first).
func (c *Client) collectEvents(ctx context.Context, namespace string, snap *Snapshot) {
	list, err := c.Clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{Limit: 500})
	if err != nil {
		return
	}
	for i := range list.Items {
		e := &list.Items[i]
		ts := e.LastTimestamp.Time
		if ts.IsZero() {
			ts = e.EventTime.Time
		}
		if ts.IsZero() {
			ts = e.CreationTimestamp.Time
		}
		snap.Events = append(snap.Events, EventInfo{
			Type:      e.Type,
			Reason:    e.Reason,
			Namespace: e.Namespace,
			Object:    e.InvolvedObject.Kind + "/" + e.InvolvedObject.Name,
			Message:   e.Message,
			Count:     e.Count,
			Age:       time.Since(ts),
		})
	}
	sort.Slice(snap.Events, func(i, j int) bool {
		return snap.Events[i].Age < snap.Events[j].Age // newest first
	})
}
