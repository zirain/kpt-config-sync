// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

const (
	// OpenTelemetry is the app label for all otel resources.
	OpenTelemetry = "opentelemetry"

	// OtelAgentName is the name of the OpenTelemetry Agent.
	OtelAgentName = "otel-agent"

	// OtelCollectorName is the name of the OpenTelemetry Collector.
	OtelCollectorName = "otel-collector"

	// OtelCollectorGooglecloud is the name of the OpenTelemetry Collector ConfigMap that contains Googlecloud exporter.
	OtelCollectorGooglecloud = "otel-collector-googlecloud"

	// OtelCollectorCustomCM is the name of the custom OpenTelemetry Collector ConfigMap.
	OtelCollectorCustomCM = "otel-collector-custom"

	// MonitoringNamespace is the Namespace used for OpenTelemetry Collector deployment.
	MonitoringNamespace = "config-management-monitoring"

	// CollectorConfigGooglecloud is the OpenTelemetry Collector configuration with
	// the googlecloud exporter.
	CollectorConfigGooglecloud = `receivers:
  opencensus:
exporters:
  prometheus:
    endpoint: :8675
    namespace: config_sync
    resource_to_telemetry_conversion:
      enabled: true
  googlecloud:
    metric:
      prefix: "custom.googleapis.com/opencensus/config_sync/"
      # The exporter would always fail at sending metric descriptor. Skipping
      # creation of metric descriptors until the error from upstream is resolved
      # The metric streaming data is not affected
      # https://github.com/GoogleCloudPlatform/opentelemetry-operations-go/issues/529
      skip_create_descriptor: true
      # resource_filters looks for metric resource attributes by prefix and converts
      # them into custom metric labels, so they become visible and can be accessed
      # under the GroupBy dropdown list in Cloud Monitoring
      resource_filters:
        - prefix: "cloud.account.id"
        - prefix: "cloud.availability.zone"
        - prefix: "cloud.platform"
        - prefix: "cloud.provider"
        - prefix: "k8s.pod.ip"
        - prefix: "k8s.pod.namespace"
        - prefix: "k8s.pod.uid"
        - prefix: "k8s.container.name"
        - prefix: "host.id"
        - prefix: "host.name"
        - prefix: "k8s.deployment.name"
        - prefix: "k8s.node.name"
    retry_on_failure:
      enabled: false
    sending_queue:
      enabled: false
  googlecloud/kubernetes:
    metric:
      prefix: "kubernetes.io/internal/addons/config_sync/"
      # skip_create_descriptor: Metrics start with 'kubernetes.io/' have already
      # got descriptors defined internally. Skip sending dupeicated metric
      # descriptors here to prevent errors or conflicts.
      skip_create_descriptor: true
      # instrumentation_library_labels: Otel Collector by default attaches
      # 'instrumentation_version' and 'instrumentation_source' labels that are
      # not specified in our Cloud Monarch definitions, thus skipping them here
      instrumentation_library_labels: false
      # create_service_timeseries: This is a recommended configuration for
      # 'service metrics' starts with 'kubernetes.io/' prefix. It uses
      # CreateTimeSeries API and has its own quotas, so that custom metric write
      # will not break this ingestion pipeline
      create_service_timeseries: true
      service_resource_labels: false
    retry_on_failure:
      enabled: false
    sending_queue:
      enabled: false
processors:
  batch:
  # resourcedetection: This processor is needed to correctly mirror resource
  # labels from OpenCensus to OpenTelemetry. We also want to keep this same
  # processor in Otel Agent configuration as the resource labels are added from
  # there
  resourcedetection:
    detectors: [env, gcp]
  filter/cloudmonitoring:
    metrics:
      include:
        match_type: regexp
        metric_names:
          - reconciler_errors
          - apply_duration_seconds
          - reconcile_duration_seconds
          - rg_reconcile_duration_seconds
          - last_sync_timestamp
          - pipeline_error_observed
          - declared_resources
          - apply_operations_total
          - resource_fights_total
          - internal_errors_total
          - kcc_resource_count
          - resource_count
          - ready_resource_count
          - cluster_scoped_resource_count
          - resource_ns_count
          - api_duration_seconds
  filter/kubernetes:
    metrics:
      include:
        match_type: regexp
        metric_names:
          - kustomize.*
          - api_duration_seconds
          - reconciler_errors
          - pipeline_error_observed
          - reconcile_duration_seconds
          - parser_duration_seconds
          - declared_resources
          - apply_operations_total
          - apply_duration_seconds
          - resource_fights_total
          - remediate_duration_seconds
          - resource_conflicts_total
          - internal_errors_total
          - rendering_count_total
          - skip_rendering_count_total
          - resource_override_count_total
          - git_sync_depth_override_count_total
          - no_ssl_verify_count_total
          - kcc_resource_count
      exclude:
        match_type: strict
        metric_names:
          - rg_reconcile_duration_seconds
          # TODO remove kcc_resource_count_total rule once Resource Group Controller
          # 1.0.9 is updated into Config Sync. This metric was unintentionally
          # included by the 'regex include' filter above and is not included in
          # our Monarch metric definitions
          - kcc_resource_count_total
  attributes/kubernetes:
    actions:
      # Remove custom configsync metric labels that are not registered with Monarch
      - key: configsync.sync.kind
        action: delete
      - key: configsync.sync.name
        action: delete
      - key: configsync.sync.namespace
        action: delete
      # Remove high cardinality configsync metric labels when sending to Monarch.
      # These labels are useful to users, but too noisy for global aggregation.
      - key: commit
        action: delete
      - key: type
        action: delete
  metricstransform/kubernetes:
    transforms:
      - include: declared_resources
        action: update
        new_name: current_declared_resources
      - include: reconciler_errors
        action: update
        new_name: last_reconciler_errors
      - include: pipeline_error_observed
        action: update
        new_name: last_pipeline_error_observed
      - include: apply_operations_total
        action: update
        new_name: apply_operations_count
      - include: resource_fights_total
        action: update
        new_name: resource_fights_count
      - include: resource_conflicts_total
        action: update
        new_name: resource_conflicts_count
      - include: internal_errors_total
        action: update
        new_name: internal_errors_count
      - include: rendering_count_total
        action: update
        new_name: rendering_count
      - include: skip_rendering_count_total
        action: update
        new_name: skip_rendering_count
      - include: resource_override_count_total
        action: update
        new_name: resource_override_count
      - include: git_sync_depth_override_count_total
        action: update
        new_name: git_sync_depth_override_count
      - include: no_ssl_verify_count_total
        action: update
        new_name: no_ssl_verify_count
extensions:
  health_check:
service:
  extensions: [health_check]
  pipelines:
    metrics/cloudmonitoring:
      receivers: [opencensus]
      processors: [batch, filter/cloudmonitoring, resourcedetection]
      exporters: [googlecloud]
    metrics/prometheus:
      receivers: [opencensus]
      processors: [batch]
      exporters: [prometheus]
    metrics/kubernetes:
      receivers: [opencensus]
      processors: [batch, filter/kubernetes, attributes/kubernetes, metricstransform/kubernetes, resourcedetection]
      exporters: [googlecloud/kubernetes]`
)
