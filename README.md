# knitfab-api-types

Types for Knitfab WebAPI

## Install

```
go get github.com/opst/knitfab-api-types
```

## Package Structure

- `data`: Types for Knitfab Data related WebAPI
- `plans`: Types for Knitfab Plan related WebAPI
- `runs`: Types for Knitfab Run related WebAPI
- `errors`: Types for error messages from Knitfab WebAPI
- `tags`: Types for Tags used from Data and Plan
- `misc`: Miscellaneous types

## Type Name Convention

Data, Run and Plan are structed by two levels, Summary and Detail.

For each of Data, Run and Plan,

- Summary: represents its identity and important status
- Detail: in addition to Summary, represents relations with other items.

## Versioning Tag

Tag in this repository shows compatibilitiy with the Knitfab version.
For example, tag `v1.3.1` in this repo is compatible Knitfab `v1.3.1`.

And, this repo may have unstable "beta" version tags (i.e., `v1.3.1-beta1`) for developing Knitfab.
