# Anthem MRF Index Parser

A high-performance, streaming parser for Anthem's Transparency in Coverage (TiC) index files.

## Overview
This tool is designed to process massive (50GB+) JSON index files with **O(1) constant memory usage**. It filters and extracts relevant data from the index files, specifically focusing on national pricing data to extract only the In-Network negotiated rates for **Anthem / Empire BlueCross BlueShield PPO** plans in **New York**.

## Key Features
* **Streaming Architecture:** Processes the HTTP stream directly through a GZIP decompressor and JSON decoder without buffering the file to disk or RAM.
* **Hybrid Filtering:** Uses a combination of high-precision Network ID and text-based discovery to ensure high recall and precision.
* **Real-time Dashboard:** Logs records processed, MB downloaded, and unique matches found in real-time.
* **Deduplication:** Tracks unique URLs in memory to prevent duplicate entries in the output.

## Project Structure
* `main.go`: Orchestrates the HTTP connection and file setup.
* `pipeline.go`: Manages the core processing loop, error handling, and logging.
* `filter.go`: Contains the business logic, target codes, and filtering rules.
* `types.go`: Defines the JSON data contracts.
* `utils.go`: Provides low-level helpers (e.g., download byte counting).

## Usage

### Prerequisites
* Go 1.21 or higher

### Running the Parser
Since the project uses a modular structure, run all files in the package:

```bash
# Run directly
go run .

# Or build binary
go build -o anthem-parser .
./anthem-parser
```

The results will be written to `output.txt` in the same directory.

## Architecture & Analysis
### Handling Memory Limitations with Large Files
The uncompressed JSON index exceeds the RAM of most standard machines so I implemented a streaming pipeline.

Instead of loading the file, I wrap the `http.Response.Body` in a `gzip.Reader`, which feeds a `json.Decoder`.
The decoder tokenizes the stream and unmarshals only one ReportingStructure object at a time.
Once processed, the object is garbage collected. This keeps memory usage flat regardless of input size.

### Output URL Analysis
The output URLs reveal the underlying data partitioning strategy of the Blue Cross ecosystem.

- The domain (mrf.bcbs.com) and date (2026-01) remain constant.
- Codes like `72A0` identify the specific product.
- The suffix _01_of_20 indicates that pricing data is sharded across multiple files.
- To build a complete dataset, we must capture every file in the _of_ sequence for our target Network IDs.

### The Role of the 'Description' Field
The description field is critical for Discovery. While structured codes are safer, the data is inconsistent. Many files lack a clear location tag but identify as "New York PPO" in the free-text description.

It's less reliable than Network IDs due to variance (e.g., "NY" vs "New York"). I used a hybrid approach: strict code matching first, followed by a normalized text matching fallback.

### Verification Strategy (EIN Lookup)
To confirm the completeness of my filter logic, I used a "Ground Truth" verification method:

- I located the EIN for large New York employers.
- I input the EIN into Anthem's MRF Search Tool for the reporting period 2026-01.
- The tool returned official JSON URLs containing the codes.
- While other niche codes may exist, the identified codes cover the vast majority of commercial PPO plans in the New York market.

## Trade-offs Made
### Simplicity vs. Fault Tolerance
- Decision: I chose a pure in-memory stream.
- Trade-off: If the network connection drops at 99%, the process must be restarted from the beginning. An assumption is made that the script is being run on a stable network connection.

### Single-Threaded vs. Concurrent
- Decision: The parsing loop runs on a single thread.
- Trade-off: I am not utilizing all CPU cores. Concurrent execution would add significant complexity with only a small performance gain since network speed is the main limiting factor.

## Time Analysis
- Time to Write: ~4 hours
    - Includes: Researching the JSON structure, identifying the specific Network IDs for NY, and refactoring from an initial Node.js prototype to this Go implementation.

- Time to Run: ~6 minutes