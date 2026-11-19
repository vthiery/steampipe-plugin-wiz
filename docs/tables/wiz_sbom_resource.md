---
title: "Steampipe Table: wiz_sbom_resource - Query Wiz SBOM Resources using SQL"
description: "Allows users to query the resources for which Wiz holds a software bill of materials (SBOM), with the number of artifacts and versions in each."
---

# Table: wiz_sbom_resource - Query Wiz SBOM Resources using SQL

Wiz builds a software bill of materials (SBOM) for the resources it scans: repository branches, container images, running containers, virtual machines, serverless functions and CI jobs. Each SBOM lists the artifacts (OS packages, code libraries, hosted technologies) found in the resource.

## Table Usage Guide

The `wiz_sbom_resource` table lists the resources for which Wiz holds an SBOM, one row per resource, with the number of artifacts and distinct versions in it. Use it to check SBOM coverage, for example that every production repository or container image has one. A resource that Wiz has not scanned, or in which it found no artifacts, does not appear.

Filter on `type` (for example `REPOSITORY_BRANCH` or `CONTAINER_IMAGE`) or `id` to let Wiz narrow the results; listing every resource can take a while in large environments.

## Examples

### Basic info
List the container images that have an SBOM, with the number of artifacts and versions in each. Change the `type` to look at another kind of resource.

```sql+postgres
select
  name,
  type,
  cloud_platform,
  artifact_count,
  version_count
from
  wiz_sbom_resource
where
  type = 'CONTAINER_IMAGE';
```

```sql+sqlite
select
  name,
  type,
  cloud_platform,
  artifact_count,
  version_count
from
  wiz_sbom_resource
where
  type = 'CONTAINER_IMAGE';
```

### Count resources with an SBOM per type
Count the resources with an SBOM for each resource type, with the total number of artifacts. This gives a quick view of SBOM coverage across repositories, images, virtual machines and the other resource types.

```sql+postgres
select
  type,
  count(*) as resources,
  sum(artifact_count) as artifacts
from
  wiz_sbom_resource
group by
  type
order by
  resources desc;
```

```sql+sqlite
select
  type,
  count(*) as resources,
  sum(artifact_count) as artifacts
from
  wiz_sbom_resource
group by
  type
order by
  resources desc;
```

### List the repository branches with an SBOM
List the repository branches Wiz holds an SBOM for. Use it to check that the repositories you expect are covered.

Repository branch names have the form `owner/repository/branch`.

```sql+postgres
select
  name,
  artifact_count,
  version_count
from
  wiz_sbom_resource
where
  type = 'REPOSITORY_BRANCH'
order by
  name;
```

```sql+sqlite
select
  name,
  artifact_count,
  version_count
from
  wiz_sbom_resource
where
  type = 'REPOSITORY_BRANCH'
order by
  name;
```

### List the container images of a project with an SBOM
List the container images of a given project that have an SBOM. This helps check SBOM coverage for the images one team or environment owns.

```sql+postgres
select
  r.name,
  r.artifact_count
from
  wiz_sbom_resource r,
  jsonb_array_elements(r.projects) as p
where
  r.type = 'CONTAINER_IMAGE'
  and p ->> 'name' = 'Production';
```

```sql+sqlite
select
  r.name,
  r.artifact_count
from
  wiz_sbom_resource r,
  json_each(r.projects) as p
where
  r.type = 'CONTAINER_IMAGE'
  and json_extract(p.value, '$.name') = 'Production';
```
