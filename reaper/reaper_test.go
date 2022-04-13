package reaper_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/maxatome/go-testdeep/td"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/baez90/kreaper/reaper"
)

func TestKreaper_Kill(t *testing.T) {
	const defaultNamespace = "default"
	t.Parallel()
	type fields struct {
		initialState corev1.PodList
		lifetime     time.Duration
		target       reaper.Target
	}
	tests := []struct {
		name          string
		fields        fields
		modifier      func(tb testing.TB, k8sClient client.Client)
		wantErr       error
		wantRemaining td.TestDeep
	}{
		{
			name: "Empty initial state",
			fields: fields{
				lifetime: 10 * time.Second,
				target:   reaper.Target("app.kubernetes.io/name=ee8dcc4d"),
			},
			wantRemaining: td.Empty(),
		},
		{
			name: "Single pod to delete",
			fields: fields{
				target:   reaper.Target("app.kubernetes.io/name=ee8dcc4d"),
				lifetime: 100 * time.Millisecond,
				initialState: corev1.PodList{
					Items: []corev1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "demo-asdf234",
								Namespace: defaultNamespace,
								Labels: map[string]string{
									"app.kubernetes.io/name": "ee8dcc4d",
								},
							},
						},
					},
				},
			},
			wantRemaining: td.Empty(),
		},
		{
			name: "Single pod to delete - delete preemptively",
			fields: fields{
				target:   reaper.Target("app.kubernetes.io/name=ee8dcc4d"),
				lifetime: 100 * time.Millisecond,
				initialState: corev1.PodList{
					Items: []corev1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "demo-asdf234",
								Namespace: defaultNamespace,
								Labels: map[string]string{
									"app.kubernetes.io/name": "ee8dcc4d",
								},
							},
						},
					},
				},
			},
			modifier: func(tb testing.TB, k8sClient client.Client) {
				tb.Helper()
				td.CmpNoError(tb, k8sClient.DeleteAllOf(context.Background(), new(corev1.Pod)))
			},
			wantRemaining: td.Empty(),
		},
		{
			name: "Single pod to delete - one should remain",
			fields: fields{
				target:   reaper.Target("app.kubernetes.io/name=ee8dcc4d"),
				lifetime: 100 * time.Millisecond,
				initialState: corev1.PodList{
					Items: []corev1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "demo-asdf234",
								Namespace: defaultNamespace,
								Labels: map[string]string{
									"app.kubernetes.io/name": "ee8dcc4d",
								},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "demo-lkjklsdf9234",
								Namespace: defaultNamespace,
								Labels: map[string]string{
									"app.kubernetes.io/name": "ef903e61",
								},
							},
						},
					},
				},
			},
			wantRemaining: td.Len(1),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			k8sClient := fake.NewClientBuilder().
				WithLists(&tt.fields.initialState).
				Build()

			k := reaper.Kreaper{
				Client:          k8sClient,
				Lifetime:        tt.fields.lifetime,
				Target:          tt.fields.target,
				TargetNamespace: defaultNamespace,
			}

			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)

			errs := make(chan error, 1)

			go func(ctx context.Context, errs chan<- error) {
				defer close(errs)
				errs <- k.Kill(ctx)
				if err := k.Kill(ctx); !errors.Is(err, tt.wantErr) {
					t.Errorf("Kill() error = %v, wantErr %v", err, tt.wantErr)
				}
			}(ctx, errs)

			if tt.modifier != nil {
				tt.modifier(t, k8sClient)
			}

			for err := range errs {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Kill() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}

			var remainingPods corev1.PodList
			if err := k8sClient.List(ctx, &remainingPods); err != nil {
				t.Fatalf("Failed to list remaining pods err = %v", err)
			}

			td.Cmp(t, remainingPods.Items, tt.wantRemaining)
		})
	}
}
