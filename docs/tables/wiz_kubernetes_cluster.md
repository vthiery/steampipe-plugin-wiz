---
title: "Steampipe Table: wiz_kubernetes_cluster - Query Wiz Kubernetes Clusters using SQL"
description: "Allows users to query the Kubernetes clusters in the Wiz Security Graph, including the deployment coverage of the Wiz runtime sensor, admission controller and connector."
---

# Table: wiz_kubernetes_cluster - Query Wiz Kubernetes Clusters using SQL

Wiz builds a Security Graph of every cloud resource it discovers. Kubernetes clusters (EKS, GKE, AKS, OpenShift and self-hosted) are part of it, whether or not a Wiz component runs inside them, along with which Wiz components are deployed to each one.

## Table Usage Guide

The `wiz_kubernetes_cluster` table lists every Kubernetes cluster in the Wiz Security Graph. Use it to check runtime protection coverage (which clusters run the Wiz sensor, and which sensor versions are end of life), Kubernetes version currency, and cluster inventory per cloud account. Properties without a dedicated column are available in the `properties` column.

## Examples

### Basic info
Get an overview of the clusters Wiz knows about, with their Kubernetes version and whether the Wiz runtime sensor is deployed. This is a good starting point to see which clusters exist and where they run.

```sql+postgres
select
  name,
  cloud_platform,
  kubernetes_flavor,
  region,
  version,
  sensor_deployment_status
from
  wiz_kubernetes_cluster;
```

```sql+sqlite
select
  name,
  cloud_platform,
  kubernetes_flavor,
  region,
  version,
  sensor_deployment_status
from
  wiz_kubernetes_cluster;
```

### List active clusters without the Wiz runtime sensor
Find the active clusters that do not run the Wiz runtime sensor. This helps identify gaps in runtime protection coverage so you can prioritize where to deploy the sensor.

```sql+postgres
select
  name,
  cloud_platform,
  subscription_external_id,
  region
from
  wiz_kubernetes_cluster
where
  status = 'Active'
  and sensor_deployment_status is distinct from 'Installed';
```

```sql+sqlite
select
  name,
  cloud_platform,
  subscription_external_id,
  region
from
  wiz_kubernetes_cluster
where
  status = 'Active'
  and coalesce(sensor_deployment_status, '') <> 'Installed';
```

### Check sensor coverage of clusters in production cloud accounts
Check the runtime sensor status of the clusters that live in cloud accounts used for production, based on the environment each account has in its Wiz projects. This helps verify that the clusters that matter most are covered, and whether their sensors are end of life.


```sql+postgres
with production_accounts as (
  select distinct
    l -> 'cloudAccount' ->> 'externalId' as account
  from
    wiz_project p,
    jsonb_array_elements(p.cloud_account_links) as l
  where
    not p.archived
    and l ->> 'environment' = 'PRODUCTION'
)
select
  c.name,
  c.subscription_external_id,
  c.sensor_deployment_status,
  c.sensor_version,
  c.sensor_is_end_of_life
from
  wiz_kubernetes_cluster c
  join production_accounts a on a.account = c.subscription_external_id
where
  c.status = 'Active';
```

```sql+sqlite
with production_accounts as (
  select distinct
    json_extract(l.value, '$.cloudAccount.externalId') as account
  from
    wiz_project p,
    json_each(p.cloud_account_links) as l
  where
    not p.archived
    and json_extract(l.value, '$.environment') = 'PRODUCTION'
)
select
  c.name,
  c.subscription_external_id,
  c.sensor_deployment_status,
  c.sensor_version,
  c.sensor_is_end_of_life
from
  wiz_kubernetes_cluster c
  join production_accounts a on a.account = c.subscription_external_id
where
  c.status = 'Active';
```

### List clusters running an end-of-life Kubernetes version
Find the clusters running a Kubernetes version that is no longer supported, next to the latest available version. This helps plan upgrades for clusters that no longer receive security fixes.

```sql+postgres
select
  name,
  cloud_platform,
  version,
  properties ->> 'latestVersion' as latest_version
from
  wiz_kubernetes_cluster
where
  is_version_end_of_life;
```

```sql+sqlite
select
  name,
  cloud_platform,
  version,
  json_extract(properties, '$.latestVersion') as latest_version
from
  wiz_kubernetes_cluster
where
  is_version_end_of_life;
```

### List the clusters of a project
List the clusters that belong to a given Wiz project, with their sensor status. This is useful to review the coverage of the clusters one team or environment owns.

```sql+postgres
select
  c.name,
  c.cloud_platform,
  c.sensor_deployment_status
from
  wiz_kubernetes_cluster c,
  jsonb_array_elements(c.projects) as p
where
  p ->> 'name' = 'Production';
```

```sql+sqlite
select
  c.name,
  c.cloud_platform,
  c.sensor_deployment_status
from
  wiz_kubernetes_cluster c,
  json_each(c.projects) as p
where
  json_extract(p.value, '$.name') = 'Production';
```
