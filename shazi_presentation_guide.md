# 🎓 Shazi's Docksmith Presentation Guide
## Build Engine & Data Models - Complete Explanation

---

## 📋 Table of Contents
1. [Project Overview](#project-overview)
2. [Faculty Q&A Preparation](#faculty-qa-preparation)
3. [Your Module: Build Engine Deep Dive](#your-module-build-engine-deep-dive)
4. [End-to-End Build Flow](#end-to-end-build-flow)
5. [Helper Functions](#helper-functions)
6. [Cache Integration](#cache-integration)
7. [Key Talking Points](#key-talking-points)

---

## 🎯 Project Overview

### What is Docksmith?

Docksmith is a **simplified Docker-like container build and runtime system** built entirely from scratch without using Docker, runc, or containerd.

**Core Features:**
- Build container images from a `Docksmithfile` (similar to Dockerfile)
- Intelligent layer caching for fast rebuilds
- Run containers with Linux process isolation using raw OS primitives
- Everything stored locally in `~/.docksmith/` with no daemon process

### System Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    CLI (main.py)                        │
│              build | images | rmi | run                 │
└────────────────────┬────────────────────────────────────┘
                     │
        ┌────────────┼────────────┐
        │            │            │
        ▼            ▼            ▼
   ┌─────────┐ ┌─────────┐ ┌──────────┐
   │ Parser  │ │ Builder │ │ Runtime  │
   │ (Shazi) │ │ (Shazi) │ │ (Pranav) │
   └─────────┘ └────┬────┘ └──────────┘
                    │
        ┌───────────┼───────────┐
        ▼           ▼           ▼
   ┌────────┐ ┌─────────┐ ┌────────┐
   │ Cache  │ │ Layers  │ │ State  │
   │(Preksha)│ │ (Shazi) │ │(Piyush)│
   └────────┘ └─────────┘ └────────┘
```

### Team Responsibilities

| Member | Module | Key Responsibilities |
|--------|--------|---------------------|
| **Piyush** | CLI & State | Command-line interface, manifest storage, image management |
| **Shazi (You)** | Build Engine & Parser | Docksmithfile parsing, layer creation, build orchestration |
| **Preksha** | Cache System | Cache key computation, hit/miss logic, cascade behavior |
| **Pranav** | Runtime | Linux namespace isolation, process execution |

---

## 💬 Faculty Q&A Preparation

### Q1: What is Docksmith and what problem does it solve?

**Answer:**
Docksmith is a minimal container build and runtime system that demonstrates core containerization concepts without relying on Docker or other existing tools. It solves the problem of understanding how container systems work internally by implementing:

- Image building from declarative instructions
- Layer-based filesystem composition
- Intelligent caching for fast rebuilds
- Process isolation using Linux primitives

### Q2: What are the six supported instructions?

**Answer:**
```
FROM <image>:<tag>      - Load base image and inherit its layers
COPY <src> <dest>       - Copy files from build context (produces layer)
RUN <command>           - Execute shell command (produces layer)
WORKDIR <path>          - Set working directory (no layer)
ENV <key>=<value>       - Set environment variable (no layer)
CMD ["exec", "arg"]     - Default container command (no layer)
```

Only `COPY` and `RUN` produce layers. The others update image configuration.

### Q3: How does the build cache work?

**Answer:**
The cache uses **deterministic SHA-256 hashing** of:
1. Previous layer digest
2. Instruction text (exact command)
3. Current WORKDIR value
4. All ENV variables (sorted alphabetically)
5. For COPY: SHA-256 hashes of all source files

**Cache Behavior:**
- **Cache HIT**: Reuse stored layer, skip execution, print `[CACHE HIT]`
- **Cache MISS**: Execute instruction, store result, print `[CACHE MISS] Xs`
- **Cascade**: Once any step misses, all subsequent steps also miss

### Q4: How do you ensure reproducible builds?

**Answer:**
Two critical techniques in `layers.py`:

1. **Sorted tar entries**: All files added to tar archives in lexicographic order
2. **Zeroed timestamps**: All file modification times set to 0

This ensures **identical inputs always produce byte-identical layer digests**.

### Q5: How does container isolation work?

**Answer:**
Uses **Linux OS primitives directly** (no Docker/runc):
- `unshare()` with `CLONE_NEWPID`, `CLONE_NEWNS`, `CLONE_NEWUTS` for namespace isolation
- `chroot()` to change root filesystem
- **Same isolation primitive** used for both:
  - `RUN` commands during build
  - `docksmith run` at runtime

This ensures files written inside containers don't appear on the host filesystem.

---

## 🏗️ Your Module: Build Engine Deep Dive

### Your Responsibilities (Shazi/Prathama)

1. **Parser (`parser.py`)**: Parse Docksmithfile line by line
2. **Builder (`builder.py`)**: Orchestrate the entire build process
3. **Layers (`layers.py`)**: Create and store reproducible tar layers
4. **Models (`models.py`)**: Define shared data structures

---

## 🔄 End-to-End Build Flow: builder.py Explained

### Key Concept: Parse → Process Instructions → Produce Layers → Create Manifest

**Entry Point:** `build_image(context_dir, name, tag, no_cache)` in `builder.py`


#### 7 Key Steps in builder.py:

1. **Parse Docksmithfile** (`builder.py` Line ~20-22)
   - Calls `parse_docksmithfile()` from `parser.py`
   - Returns list of validated `Instruction` objects
   - Ensures first instruction is `FROM`

2. **Initialize Build State** (`builder.py` Line ~24-31)
   - `layers = []` - accumulates LayerEntry objects
   - `env_dict = {}` - tracks ENV variables
   - `workdir = ""` - tracks current WORKDIR
   - `cmd = []` - stores CMD array
   - `prev_digest = None` - for cache key computation
   - `cache = CacheManager(no_cache)` - Preksha's cache system

3. **Process FROM Instruction** (`builder.py` Line ~36-47)
   - Loads base image manifest using `load_manifest()` from `state.py`
   - Inherits base layers: `layers = list(base_manifest.layers)`
   - Inherits ENV, WORKDIR, CMD from base image
   - Sets `prev_digest = base_manifest.digest` for cache chain

4. **Process State-Only Instructions** (WORKDIR, ENV, CMD)
   - **WORKDIR**: Updates `workdir` variable (no layer produced)
   - **ENV**: Adds to `env_dict` using `parse_env_args()` from `parser.py`
   - **CMD**: Stores command array using `parse_cmd_args()` from `parser.py`
   - These only update image config, don't create layers

5. **Process COPY Instruction** (`builder.py` Line ~60-85)
   - Computes cache key: `hash_copy_sources()` from `layers.py` hashes all source files
   - Calls `cache.lookup()` to check for cache hit
   - **Cache HIT**: Reuses existing layer digest, prints `[CACHE HIT]`
   - **Cache MISS**: 
     - Calls `_execute_copy()` helper (`builder.py` Line ~120-150)
     - Assembles current filesystem using `_assemble_rootfs()`
     - Resolves glob patterns, copies files to destination
     - Creates delta tar using `create_delta_tar()` from `layers.py`
     - Stores layer using `store_layer()` from `layers.py`
     - Updates cache using `cache.store()`
     - Prints `[CACHE MISS] Xs` with timing

6. **Process RUN Instruction** (`builder.py` Line ~87-110)
   - Computes cache key (no copy_hashes, just instruction + state)
   - Calls `cache.lookup()` to check for cache hit
   - **Cache HIT**: Reuses existing layer digest
   - **Cache MISS**:
     - Calls `_execute_run()` helper (`builder.py` Line ~152-175)
     - Assembles current filesystem
     - Takes filesystem snapshot using `snapshot_filesystem()` from `layers.py`
     - Executes command using `isolate_and_run()` from `runtime.py` (Pranav's module)
     - Takes another snapshot, computes delta using `compute_delta_paths()` from `layers.py`
     - Creates tar of only changed files
     - Stores layer and updates cache

7. **Create and Save Manifest** (`builder.py` Line ~112-125)
   - Builds `ImageConfig` from accumulated `env_dict`, `cmd`, `workdir`
   - Creates `ImageManifest` with name, tag, config, and all layers
   - Calls `save_manifest()` from `state.py` (Piyush's module)
   - Manifest digest computed as SHA-256 of serialized JSON
   - Saved to `~/.docksmith/images/<name>_<tag>.json`

---

## 🛠️ Helper Functions in builder.py

### `_assemble_rootfs(layers, target_dir)` (`builder.py` Line ~118)
- Extracts all layer tar files in order into target directory
- Later layers overwrite earlier ones (union filesystem behavior)
- Used by both `_execute_copy()` and `_execute_run()`

### `_execute_copy(src_pattern, dest, context_dir, current_layers, workdir)` (`builder.py` Line ~120)
- Assembles current filesystem state
- Resolves glob patterns (`*`, `**`, or `.`)
- Copies matched files to destination
- Creates tar of **only copied files** (delta)
- Returns layer digest

### `_execute_run(command, current_layers, env_dict, workdir, isolate_fn)` (`builder.py` Line ~152)
- Assembles current filesystem
- Takes snapshot before execution
- Executes command in isolated environment (Pranav's `isolate_and_run()`)
- Takes snapshot after execution
- Computes delta (only changed files)
- Creates tar of **only changed files**
- Returns layer digest

### `_serialize_env(env_dict)` (`builder.py` Line ~115)
- Converts env dict to deterministic string: `"KEY1=val1&KEY2=val2"`
- Sorted alphabetically for cache key consistency

---

## 🔐 Cache Integration

### How Cache is Wired in builder.py

**Before COPY/RUN execution:**
```python
cached_digest = cache.lookup(
    prev_digest=prev_digest,
    instruction_text="COPY . /app",
    workdir=workdir,
    env_serialized=_serialize_env(env_dict),
    copy_hashes=hash_copy_sources(...)  # COPY only
)
```

**After execution (on cache miss):**
```python
cache.store(
    prev_digest=prev_digest,
    instruction_text="COPY . /app",
    workdir=workdir,
    env_serialized=_serialize_env(env_dict),
    copy_hashes=hash_copy_sources(...),
    result_digest=digest
)
```

**Cascade behavior:** Once `cache.store()` is called, `cache._force_miss = True`, making all subsequent lookups return `None`

---

## 🎤 Key Talking Points

### "The builder orchestrates all modules into an actual image build pipeline"

- **Parser** provides structured instructions
- **Cache manager** (Preksha) decides hit/miss before execution
- **Layers module** creates reproducible tars
- **Runtime module** (Pranav) executes RUN commands in isolation
- **State module** (Piyush) persists the final manifest

### "COPY and RUN produce layers through delta computation"

- **COPY**: Only files being copied become the layer (not entire filesystem)
- **RUN**: Filesystem snapshot before/after, only changes become the layer
- This keeps layers small and efficient

### "Cache logic is wired through deterministic key computation"

- Every input that affects output is hashed into cache key
- Same inputs = same cache key = same layer digest
- Cache miss triggers cascade flag
- Reproducible builds ensure same inputs = same digest

---

## 📁 File Structure Reference

```
docksmith/
├── models.py          # Data structures (ImageManifest, LayerEntry, ImageConfig)
├── parser.py          # Docksmithfile parsing & validation
├── builder.py         # Build orchestration & layer execution ← YOUR MAIN FILE
├── layers.py          # Tar creation, storage, extraction, delta computation
├── cache.py           # Cache key computation, lookup, store (Preksha)
├── runtime.py         # Process isolation (Pranav)
├── state.py           # Manifest I/O (Piyush)
└── main.py            # CLI entry point (Piyush)
```

---

## ✅ Summary Checklist

Before your presentation, make sure you can explain:

- [ ] The 7 key steps in `build_image()` function
- [ ] How FROM loads and inherits base layers
- [ ] How WORKDIR/ENV/CMD update state without producing layers
- [ ] How COPY produces a layer (glob → copy → delta tar → store)
- [ ] How RUN produces a layer (snapshot → execute → delta → store)
- [ ] The 4 helper functions and their purposes
- [ ] How cache lookup happens before COPY/RUN
- [ ] How cache store happens after execution
- [ ] How cascade behavior works (`cache._force_miss = True`)

---

**Good luck with your presentation, Shazi! 🚀**
