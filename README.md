# kube-events-exporter
[![Build Status](https://travis-ci.org/rhobs/kube-events-exporter.svg?branch=master)](https://travis-ci.org/rhobs/kube-events-exporter)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

**⚠️** This project is still under development and has not been proven stable yet.

## Overview

Kubernetes Events provide an insight into the cluster, we want to leverage that
by aggregating and exposing metrics about these events. It will be left to the
user to choose which Events should be aggregated and exposed by the exporter.

## Design

Please refer to the [design document](./docs/design-document.md).

## Metrics

The exporter exposes metrics over two different HTTP servers. One exposing
metrics about Kubernetes Events and another one metrics about the exporter
health.

Note that the listening addresses of the servers can be configured via flags..

From the information gathered on the Events, the expoter expose the following
metric:

```
kube_events_total{type=””, involved_object_namespace=””, involved_object_kind=””, reason=””}
```

This metric represent the number of Event that were emitted since the exporter
started running in the cluster.

Part of this metric, it is possible to get information about the Event type
(Normal, Warning, ...), its reason and the object involved.

## Cardinality

The cardinality of the metrics exposed by the default configuration of the
exporter is reasonable. After running it for two months in an OpenShift cluster
with around 60 namespaces, the exporter only generated around 500 series.

However, the cardinality tends to grow exponentially along with the size of the
cluster. Thus, we added mechanisms to reduce the amount of series generated and
make sure that all the labels are bounded.

The exporter has flags filtering some Events to reduce the number of generated
series.


- --event-types : List of allowed Event types. Defaults to all types.
- --involved-object-api-groups : List of allowed Event involved object API groups. Defaults to all API groups.
- --involved-object-namespaces : List of allowed Event involved object namespaces. Defaults to all namespaces.
- --reporting-controllers : List of controllers allowed to report Event. Defaults to all controllers.

For example, if we want to only expose metrics about Warning Events involving
v1 API group objects reported by the kubelet in the default namespace. We
would have the following flags set:
```
--event-types=Warning
--involved-object-api-groups=v1
--involved-object-namespaces=default
--reporting-controllers=kubelet
```

## Prerequisites

The exporter supports Kubernetes clusters starting from v1.17.0+.

> Note: the exporter might work in previous versions but there is no guarantee.

## Quickstart

To try the exporter in a Kubernetes cluster, you can run the following command:

```sh
kubectl create -f manifests/
```

This will create the most basic configuration for the exporter to run in the
cluster and expose metrics to Prometheus.

## Roadmap

* Expose more labels
* Expose metrics over HTTPS
* Provide alerting/recording rules
