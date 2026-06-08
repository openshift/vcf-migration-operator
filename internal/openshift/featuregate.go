package openshift

import (
	"context"
	"fmt"
	"os"
	"time"

	configclient "github.com/openshift/client-go/config/clientset/versioned"
	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	"github.com/openshift/library-go/pkg/operator/configobserver/featuregates"
	"github.com/openshift/library-go/pkg/operator/events"
	"k8s.io/client-go/rest"
	"k8s.io/utils/clock"
)

const (
	featureGateInitTimeout        = time.Minute
	missingVersionMarker          = "0.0.1-snapshot"
	releaseVersionEnvVariableName = "RELEASE_VERSION"
)

// SetupFeatureGateAccessor starts the standard library-go feature-gate accessor
// and waits for the initial feature-gate state before returning.
func SetupFeatureGateAccessor(ctx context.Context, restConfig *rest.Config) (featuregates.FeatureGateAccess, error) {
	configClient, err := configclient.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("creating config client for feature gate accessor: %w", err)
	}

	configInformers := configinformers.NewSharedInformerFactory(configClient, 10*time.Minute)
	featureGateAccessor := featuregates.NewFeatureGateAccess(
		getReleaseVersion(),
		missingVersionMarker,
		configInformers.Config().V1().ClusterVersions(),
		configInformers.Config().V1().FeatureGates(),
		events.NewLoggingEventRecorder("vcf-migration-operator", clock.RealClock{}),
	)

	go featureGateAccessor.Run(ctx)
	go configInformers.Start(ctx.Done())

	select {
	case <-featureGateAccessor.InitialFeatureGatesObserved():
		return featureGateAccessor, nil
	case <-time.After(featureGateInitTimeout):
		return nil, fmt.Errorf("timed out waiting for feature gate detection")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func getReleaseVersion() string {
	releaseVersion := os.Getenv(releaseVersionEnvVariableName)
	if releaseVersion == "" {
		return missingVersionMarker
	}

	return releaseVersion
}
