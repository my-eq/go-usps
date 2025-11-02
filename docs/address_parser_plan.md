# USPS Address Parser Design Plan

## Reference Materials
- [USPS Publication 28: Postal Addressing Standards](https://pe.usps.com/archive/pdf/DMMArchive20050109/pub28.pdf)
- [USPS Addressing Standards (DMM 602)](https://pe.usps.com/text/dmm300/602.htm)
- [USPS AIS Product Technical Guides](https://postalpro.usps.com/address-quality/ais-products)

These official USPS resources provide authoritative rules for component ordering, abbreviations, secondary unit designators, and ZIP Code usage. They should be treated as the primary source when resolving ambiguities in user input.

## Goals
1. **Normalize free-form addresses** into the structured fields required by `models.AddressRequest` while preserving enough context to offer corrective feedback.
2. **Offer rich diagnostics** that pinpoint missing components, conflicting data, or violations of USPS Publication 28 formatting rules.
3. **Deliver extensible architecture** that supports future enhancements (e.g., international addresses) without rewriting core logic.

## Architectural Overview
```
┌─────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  Tokenizer  │ -> │ Normalizer   │ -> │ Validator    │ -> │ Formatter    │
└─────────────┘    └──────────────┘    └──────────────┘    └──────────────┘
        │                   │                   │                   │
        ▼                   ▼                   ▼                   ▼
  Token stream        Canonical tokens    Structured address   USPS request
                                         + diagnostics         payload
```

### Tokenizer
- Implements a deterministic finite automaton that classifies lexemes (house numbers, directional indicators, street types, secondary units, cities, states, ZIPs).
- Uses USPS Publication 28 appendices to seed lookup tables for standard abbreviations and directional variants.
- Handles punctuation, case normalization, and whitespace collapse up front to simplify downstream stages.

### Normalizer
- Applies USPS abbreviation rules (Publication 28, Appendix C) and converts directional words to their standardized forms.
- Maintains provenance metadata (token source spans) so validation errors can highlight precise locations in the user input.
- Resolves ambiguous tokens with heuristics backed by frequency dictionaries (e.g., distinguishing "NW" as directional versus part of a street name) and defers to diagnostics when ambiguity remains unresolved.

### Validator
- Enforces component ordering and presence requirements per Publication 28, Chapters 1–3.
- Cross-validates state and ZIP combinations using USPS AIS reference data; emits diagnostics when mismatches occur.
- Detects mutually exclusive constructs (e.g., PO Box with street delivery) and flags them with actionable remediation guidance.

### Formatter
- Maps the validated structure to `models.AddressRequest`, ensuring maximum field lengths are respected.
- Produces suggestions for correcting input (e.g., "Add a 4-digit ZIP Code extension" or "Replace 'Apartment' with 'APT'") based on the diagnostics metadata.

## Data Structures & Patterns
- **Trie-based lexicons** for street suffixes, directional indicators, and secondary unit designators enable efficient prefix matching and normalization.
- **Finite state machines** drive tokenization and disambiguation logic while keeping the lexer deterministic.
- **Pipeline pattern** maintains separation of concerns across stages; each stage emits structured artifacts and diagnostics.
- **Builder pattern** assembles the final `ParsedAddress` while preserving immutability of intermediate representations.
- **Strategy pattern** supports pluggable heuristics (e.g., rural route handling vs. PO Box parsing) without branching complexity.

## Error Handling & Diagnostics
- Diagnostics include severity, user-facing message, offending span, and remediation hints referencing Publication 28 sections for self-service fixes.
- Aggregated diagnostics allow UI clients to present multiple issues simultaneously (e.g., missing city and malformed ZIP).
- Provide structured machine-readable codes to support localization and automated remediation workflows.

## Testing Strategy
- **Table-driven unit tests** covering residential addresses, PO Boxes, rural routes, military mail, intersections, and secondary units.
- **Fuzz testing** to uncover tokenizer edge cases (e.g., mixed Unicode punctuation, misspellings).
- **Golden files** comparing expected normalization outputs for curated edge cases derived from Publication 28 examples.
- **Property-based checks** ensuring round-trip formatting (normalized output, then re-parse) maintains idempotence.

## Integration Points
- Expose a `parser.Parse` entry point returning `(ParsedAddress, []Diagnostic)`.
- Provide helpers to convert `ParsedAddress` into `models.AddressRequest` and to serialize diagnostics for UI consumption.
- Supply configuration hooks for injecting AIS datasets or custom heuristics during initialization.

## Future Enhancements
- Integrate USPS APIs (e.g., Address Validation, AIS) for runtime verification once credentials are available.
- Support localized parsing profiles for territories (e.g., Puerto Rico addressing nuances) and eventually international formats.
- Offer telemetry hooks to log parsing failures for iterative improvement of heuristics.
