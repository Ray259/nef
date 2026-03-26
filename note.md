# Developer Handover Notes
**Author**: Gemini
**Date**: March 25, 2026
**Feature**: NEF UE Identifier (3GPP TS 29.522 / TS 29.503)

This document contains technical concerns, technical debt, and context regarding the implementation of the `3gpp-ueid` feature for the next developer.

## 1. Context: NEF UE Identifier
The NEF UE Identifier feature (Northbound `3gpp-ueid`) allows Application Functions (AF) to request internal 5G core identifiers (like GPSI) without exposing sensitive identifiers like the SUPI. The NEF acts as a security boundary: it queries the UDM (Southbound `Nudm_SDM`), receives the SUPI and GPSI, but explicitly scrubs the SUPI before returning the response to the AF.

## 2. Technical Concerns & Debt

### A. Incomplete 3GPP AF Query Support (IP / MAC Translation)
- **Problem**: According to 3GPP TS 29.522, an AF can query a UE Identifier using either a `GPSI`, an `IPv4/IPv6 Address`, or a `MAC Address`. Currently, our `RetrieveUEId` processor only supports direct GPSI lookups.
- **Impact**: If an AF submits a request containing only an `ipAddr`, the current implementation rejects it with a "Missing Gpsi" `ProblemDetails` error.
- **Next Steps**: To fully comply with 3GPP specifications for IP-based or MAC-based identifier retrieval, the NEF needs to implement a secondary resolution mechanism. When `ipAddr` is provided instead of `gpsi`, the NEF should first query the **BSF (Binding Support Function)** or **UDR (Unified Data Repository)** to translate the IP address to a GPSI/SUPI, and then proceed to query the UDM.

### B. Temporary Local Models (openapi v1.2.3 limitations)
- **Problem**: The project is currently using `github.com/free5gc/openapi v1.2.3`. This specific version does not yet contain the generated Golang models for the Northbound `3gpp-ueid` API (`UeIdReq`, `UeIdInfo`, `UeIdResult`).
- **Workaround**: To maintain TS 29.522 JSON payload compliance, these models were manually declared in `internal/sbi/processor/models_ueid.go`.
- **Next Steps**: This is technical debt. When the upstream `free5gc/openapi` dependency is updated to a newer release that includes the UE ID models, `models_ueid.go` should be deleted, and all imports should be updated to use the official `models.UeIdReq` and `models.UeIdInfo`.

### C. Configuration Validation Constraint (Fixed but worth noting)
- **Context**: In `pkg/factory/config.go`, the `ServiceList` validation was previously hardcoded to only allow `ServiceNefPfd` and `ServiceNefOam`.
- **Action Taken**: This was patched to include `ServiceNefUeId`, `ServiceTraffInflu`, and `ServiceNefCallback` to prevent the application from crashing on startup when these services are enabled in `nefcfg.yaml`.
- **Next Steps**: If any additional Northbound APIs are implemented in the future, developers must remember to add the new service name to the `switch` statement in `config.Validate()` located in `pkg/factory/config.go`.

## 3. Testing Context
A mock-based test suite has been added in `internal/sbi/processor/ueid_test.go` using `gock`. It successfully validates:
1. Normal GPSI translation via UDM SDM.
2. Missing GPSI rejection.
3. Successful exclusion of SUPI from the returned AF payload (verification of the NEF trust boundary).

To run the related tests:
```bash
go test -v ./internal/sbi/processor/...
```
