# Graph Report - /Users/EVT/Developer/pet_projects/ogn-client  (2026-04-19)

## Corpus Check
- Corpus is ~5,481 words - fits in a single context window. You may not need a graph.

## Summary
- 99 nodes · 133 edges · 10 communities detected
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 21 edges (avg confidence: 0.79)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Project Overview & Rationale|Project Overview & Rationale]]
- [[_COMMUNITY_Public API Surface|Public API Surface]]
- [[_COMMUNITY_Parser Internals|Parser Internals]]
- [[_COMMUNITY_TCP Client Lifecycle|TCP Client Lifecycle]]
- [[_COMMUNITY_APRS Filters & Geo Utilities|APRS Filters & Geo Utilities]]
- [[_COMMUNITY_Parser Test Suite|Parser Test Suite]]
- [[_COMMUNITY_Filter Builders|Filter Builders]]
- [[_COMMUNITY_Device Database (DDB)|Device Database (DDB)]]
- [[_COMMUNITY_CheapRuler Geo Math|CheapRuler Geo Math]]
- [[_COMMUNITY_ConnectDisconnect Doc|Connect/Disconnect Doc]]

## God Nodes (most connected - your core abstractions)
1. `Parse()` - 10 edges
2. `ogn-client (Go library)` - 9 edges
3. `Client` - 8 edges
4. `ParsePosition()` - 7 edges
5. `client package` - 7 edges
6. `parseStatus()` - 6 edges
7. `parser.Parse` - 6 edges
8. `PositionMessage` - 6 edges
9. `Server-side filtering (reduces traffic)` - 6 edges
10. `main()` - 5 edges

## Surprising Connections (you probably didn't know these)
- `Parse()` --calls--> `New()`  [INFERRED]
  parser/parser.go → client/client.go
- `main()` --calls--> `Parse()`  [INFERRED]
  examples/main.go → parser/parser.go
- `ParsePosition()` --calls--> `New()`  [INFERRED]
  parser/parser.go → client/client.go
- `main()` --calls--> `RangeFilter()`  [INFERRED]
  examples/main.go → client/filter.go
- `main()` --calls--> `New()`  [INFERRED]
  examples/main.go → client/client.go

## Hyperedges (group relationships)
- **APRS filter builders for server-side filtering** — readme_range_filter, readme_area_filter, readme_prefix_filter, readme_budlist_filter, readme_type_filter, readme_combine_filters, readme_server_side_filtering [EXTRACTED 0.95]
- **parser.Parse returns one of four message types** — readme_parse, readme_position_message, readme_status_message, readme_server_message, readme_comment_message [EXTRACTED 0.95]
- **TCP client connection pattern (ports, keep-alive, auto-reconnect)** — readme_client_run, readme_port_full_feed, readme_port_filtered, readme_default_keepalive, readme_auto_reconnect, readme_auto_port_selection [EXTRACTED 0.90]

## Communities

### Community 0 - "Project Overview & Rationale"
Cohesion: 0.13
Nodes (19): CLAUDE.md project context, Testing (go test ./parser -v), APRS Protocol, Auto port selection based on filter, Auto-reconnect (5s backoff), client package, ddb package (device database), DefaultHost aprs.glidernet.org (+11 more)

### Community 1 - "Public API Surface"
Cohesion: 0.13
Nodes (17): BeaconType APRS destination map (OGFLR/OGTRK/OGNFNT/OGNSDR), Public API summary, AircraftType codes (0-15), BeaconType (flarm/tracker/receiver/etc.), client.New, Client.Run, Client.RunWithTimedCallback, CommentMessage (+9 more)

### Community 2 - "Parser Internals"
Cohesion: 0.21
Nodes (13): BeaconType, CommentMessage, createTimestamp(), MessageType, parseFloat(), parseInt(), parseOGNExtension(), ParsePosition() (+5 more)

### Community 3 - "TCP Client Lifecycle"
Cohesion: 0.26
Nodes (4): Client, New(), RangeFilter(), main()

### Community 4 - "APRS Filters & Geo Utilities"
Cohesion: 0.31
Nodes (9): AreaFilter, BudlistFilter, CheapRuler (fast distance/bearing <500km), CombineFilters, NormalizedQuality, PrefixFilter, RangeFilter, Server-side filtering (reduces traffic) (+1 more)

### Community 5 - "Parser Test Suite"
Cohesion: 0.43
Nodes (7): Parse(), mustTime(), TestCreateTimestamp(), TestParseComment(), TestParsePosition(), TestParseServerComment(), TestParseStatus()

### Community 6 - "Filter Builders"
Cohesion: 0.33
Nodes (0): 

### Community 7 - "Device Database (DDB)"
Cohesion: 0.4
Nodes (2): Device, response

### Community 8 - "CheapRuler Geo Math"
Cohesion: 0.4
Nodes (3): CheapRuler, NewCheapRuler(), TestCheapRulerDistance()

### Community 9 - "Connect/Disconnect Doc"
Cohesion: 1.0
Nodes (1): Client.Connect/Disconnect

## Knowledge Gaps
- **29 isolated node(s):** `Device`, `response`, `MessageType`, `BeaconType`, `StatusMessage` (+24 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `Connect/Disconnect Doc`** (1 nodes): `Client.Connect/Disconnect`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Parse()` connect `Parser Test Suite` to `Parser Internals`, `TCP Client Lifecycle`?**
  _High betweenness centrality (0.114) - this node is a cross-community bridge._
- **Why does `main()` connect `TCP Client Lifecycle` to `Parser Test Suite`?**
  _High betweenness centrality (0.095) - this node is a cross-community bridge._
- **Why does `RangeFilter()` connect `TCP Client Lifecycle` to `Filter Builders`?**
  _High betweenness centrality (0.052) - this node is a cross-community bridge._
- **Are the 7 inferred relationships involving `Parse()` (e.g. with `TestParsePosition()` and `TestParseStatus()`) actually correct?**
  _`Parse()` has 7 INFERRED edges - model-reasoned connections that need verification._
- **What connects `Device`, `response`, `MessageType` to the rest of the system?**
  _29 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Project Overview & Rationale` be split into smaller, more focused modules?**
  _Cohesion score 0.13 - nodes in this community are weakly interconnected._
- **Should `Public API Surface` be split into smaller, more focused modules?**
  _Cohesion score 0.13 - nodes in this community are weakly interconnected._