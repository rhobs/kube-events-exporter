# Design document: Kubernetes events aggregator and exporter

## Summary

Kubernetes Events provide an insight into the cluster, we want to leverage that
by aggregating and exposing metrics about these events. It will be left to the
user to choose which Events should be aggregated and exposed by the exporter.

## Motivation

This exporter aims to improve users' experience with Kubernetes Events by using
Prometheus. Currently if Kubernetes users want to use information from Events
they can only access them tediously from the command line for a limited period
of time.

To improve that, the exporter will expose metrics out of Kubernetes Events to
Prometheus in order to see metrics about Events in a timeline and correlate
them with existing metrics. Another goal is to create info alerts based on
warning Events and existing metrics to make it easier to detect important
changes in the cluster.

## Proposal

Users usually interact with Kubernetes events by running kubectl get events.
Example of an output for “kubectl get events” in a particular namespace: 

```
LAST SEEN   TYPE      REASON      OBJECT        MESSAGE

<unknown>   Normal    Scheduled   pod/example   Successfully assigned lcosic/example to ip-10-0-162-171.us-east-2.compute.internal

81s         Normal    Pulling     pod/example   Pulling image "openshift/hello-openshift:12346"

81s         Warning   Failed      pod/example   Failed to pull image "openshift/hello-openshift:12346": rpc error: code = Unknown desc = Error reading manifest 12346 in docker.io/openshift/hello-openshift: manifest unknown: manifest unknown

81s         Warning   Failed      pod/example   Error: ErrImagePull

66s         Normal    BackOff     pod/example   Back-off pulling image "openshift/hello-openshift:12346"

54s         Warning   Failed      pod/example   Error: ImagePullBackOff
```

From this, users are able to get more in depth information about changes
happening in the cluster. However, they are only able to know this by running
the kubectl get events command manually. To make it easier for users to detect
changes in their cluster, this exporter will allow users to correlate events
with other metrics and create alerts out of them. To do so, it will expose
Prometheus metrics from Kubernetes Events.

In order to create meaningful metrics and labels, the exporter will look for
the following Event fields:

- **Type**: Type of this event (Normal, Warning), new types could be added in
            the future
- **InvolvedObject**:  The object that this event is about.
    - **Namespace**: Namespace of the referent.
    - **Kind**: Kind of the referent.
    - **Name**:  of the referent.
- **Reason**: This should be a short, machine understandable string that gives
              the reason for the transition into the object's current status.

| Labels                    | Bounded | Limited | Cardinality | Exposed by default |
|---------------------------|---------|---------|-------------|--------------------|
| type                      | Yes     | Yes     | Low         | Yes                |
| involved_object_namespace | No      | Yes     | Moderate    | Yes                |
| involved_object_kind      | No      | Yes     | Moderate    | Yes                |
| involved_object_name      | No      | No      | High        | No                 |
| reason                    | No      | Yes     | Moderate    | Yes                |

Since some clusters have a lot of namespaces, the involved_object_namespace
label can be at the origin of a lot of series. Thus, in order to bound the
values this label can take, the exporter will be configurable to only expose
metrics from Events involving objects from predefined namespaces.

Most Kubernetes objects create Events in order to reflect state changes, part
of which are Custom Resources. Considering that, the involved_object_kind label
can take an unbound amount of values. In order to limit the number of possible
values, the default configuration of the exporter will only expose metrics from
Events involving objects from Kube API groups. It will then be left to the
users to specify which additional API groups they want to expose.

The name field of the involved object being an id, it is considered as unbound
and high cardinality. Thus, the involved_object_name label will not be exposed
by the default configuration of the exporter. However, users will be able to
allow it by passing adequate arguments to the exporter.

The Event’s reason field is a standalone string; however, controllers bring
predefined reasons. Since any controller can bring new reasons, the reason
label can take an unbounded amount of values. In order to limit the number of
reasons, the exporter will by default only expose reasons from Events emitted
by Kubernetes native controllers since they bring a limited number of reasons.
It will then be left opt-in to add controllers from which Event’s reasons
should be exposed.

Out of these information the exporter will expose the following metrics:

```
kube_events_total{type=””, involved_object_namespace=””, involved_object_kind=””, reason=””}
```

### Goals

The project is split into 2 phases. The first one main goals are to create
an exporter exposing metrics from warning level Kubernetes Events and to
provide integration for deploying into Kubernetes clusters. Whilst the second
is driven by the exposition of metrics from normal level Events.

#### Phase 1

Exporter:

- Expose Prometheus metrics from Kubernetes warning level Events.
- Expose RED metrics about the exporter.
- Allow granularity via arguments.
- Add basic logs and allow to set log levels (e.g. debug).
- Add unit and end-to-end tests.

Integration:

- Add Kubernetes Deployment and PodMonitor via jsonnet.
- Ensure that labels enabled by default are bound.
- Provide libsonnet that could be imported.
- Create predefined alerting rules mixin based on the exporter’s metrics.

#### Phase 2

- Expose Prometheus metrics from Kubernetes normal level Events.
- Improve granularity by adding the involved component’s name as a label.
- Add new labels to get more information about the events. e.g:
    - **source_component**: Component from which the event is generated.
    - **reporting_component**: Name of the controller that emitted this Event,
                               e.g. `kubernetes.io/kubelet`.
    - **reporting_Instance**: ID of the controller instance, e.g. `kubelet-xyzf`

### User Stories

#### Story 1

As a user I would like to be able to see Kubernetes Events as metrics.

#### Story 2

As a user I’d like to be able to see metrics related to the exporter's health.

#### Story 3

As a user I’d like to be able to filter the metrics scraped by the exporter by
passing args.

#### Story 4

As a user I’d like to be able to easily deploy the exporter inside of my
Kubernetes cluster.

#### Story 5

As a user I’d like to deploy the exporter in a way it is discoverable and
scrapable by Prometheus.

#### Story 6

As a user I’d like to be able to deploy predefined alerting rules based on the
exporter’s metrics.

#### Story 7

As a user I’d like to see Events related metrics in my Kubernetes cluster.

### Implementation Details

In order to efficiently watch for Kubernetes Events, the exporter will use
Reflector from the client-go library. This reflector will use a ListWatch to be
able to Watch and List Events and a Store to store the changes.

> Note: some Events are grouped using records. In that case, already existing
Events won’t be recreated but their count / series.count will be updated by a
PATCH request.

To create Prometheus metrics out of these events, the exporter will use the
client_golang library. A custom Collector will be created to list the different
metrics that should be collected and then added to a Prometheus Registry to be
exposed by the exporter.
