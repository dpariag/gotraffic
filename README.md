# Gotraffic

Gotraffic is a tool for testing networking devices using realistic traffic patterns.
Gotraffic starts with a user-specified collection of capture (pcap) files, and is able to generate
realistic traffic by mixing and matching those capture files. This is much more than tcpreplay. For example, given a single YouTube capture, gotraffic can simulate a viral video event by replaying thousands of concurrent copies of the capture. Each copy will originate from a unique IP address, but server IPs, packet payloads, and inter packet gaps are fully preserved. Add a Facebook capture and a few Twitter captures and all of sudden we're approaching a rich traffic mix.

Gotraffic provides detailed measurements of bandwidth, packet counts, latency and much more.
There's a simple web-based UI for visualizing traffic rates, and a simple REST API for retrieving
statistics.

#### Using gotraffic

