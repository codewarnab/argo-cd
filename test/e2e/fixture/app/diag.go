package app

import (
	"context"
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/argoproj/argo-cd/v3/test/e2e/fixture"
)

// dumpTimeoutDiagnostics logs what the cluster and the Application look like when an
// expectation times out. EXPERIMENT ONLY: it helps find why an app stays Progressing.
func (c *Consequences) dumpTimeoutDiagnostics() {
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	a, err := c.get()
	if err != nil {
		log.Warnf("E2E-DIAG %s cannot get app: %v", now, err)
		return
	}
	st, _ := json.Marshal(map[string]any{
		"reconciledAt": a.Status.ReconciledAt, "health": a.Status.Health,
		"sync": a.Status.Sync.Status, "resources": a.Status.Resources,
		"operationState": a.Status.OperationState,
	})
	log.Infof("E2E-DIAG %s app status: %s", now, st)
	ns := a.Spec.Destination.Namespace
	if deps, err := fixture.KubeClientset.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{}); err == nil {
		for _, d := range deps.Items {
			b, _ := json.Marshal(map[string]any{"name": d.Name, "generation": d.Generation, "rv": d.ResourceVersion, "status": d.Status})
			log.Infof("E2E-DIAG deployment: %s", b)
		}
	}
	if pods, err := fixture.KubeClientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{}); err == nil {
		for _, p := range pods.Items {
			b, _ := json.Marshal(map[string]any{"name": p.Name, "node": p.Spec.NodeName, "phase": p.Status.Phase, "conditions": p.Status.Conditions, "containers": p.Status.ContainerStatuses})
			log.Infof("E2E-DIAG pod: %s", b)
		}
	}
	if evs, err := fixture.KubeClientset.CoreV1().Events(ns).List(ctx, metav1.ListOptions{}); err == nil {
		for _, e := range evs.Items {
			log.Infof("E2E-DIAG event: %s %s/%s %s: %s (first %s last %s x%d)", e.Type, e.InvolvedObject.Kind, e.InvolvedObject.Name, e.Reason, e.Message, e.FirstTimestamp.UTC().Format(time.RFC3339), e.LastTimestamp.UTC().Format(time.RFC3339), e.Count)
		}
	}
}
