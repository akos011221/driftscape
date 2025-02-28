<p align="center">
  <img src="https://via.placeholder.com/150x50.png?text=DriftScape" alt="DriftScape Logo" />
</p>

<h1 align="center">DriftScape</h1>

<p align="center">
  <strong>A Procedural Exploration Adventure Powered by Go & Kubernetes</strong>
</p>

<p align="center">
  <a href="https://github.com/orbanakos2312/driftscape/actions"><img src="https://img.shields.io/github/workflow/status/orbanakos2312/driftscape/build?label=Build" alt="Build Status"></a>
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.21-blue" alt="Go Version"></a>
  <a href="https://kubernetes.io"><img src="https://img.shields.io/badge/Kubernetes-OKE-brightgreen" alt="Kubernetes"></a>
  <a href="https://redis.io"><img src="https://img.shields.io/badge/Redis-Enabled-red" alt="Redis"></a>
</p>

---

## 🌍 Overview

**DriftScape** is an exploration game where you roam a grid-based world that grows with every step. Start at `(0,0)`, type `move north`, and watch new regions—like sprawling forests or windswept plains—emerge as Kubernetes pods. Built with Go for lightning-fast generation, a Coordinator tracks your journey, and Redis keeps your world alive across restarts. It’s a minimalist adventure fused with cutting-edge tech, running on Oracle Kubernetes Engine (OKE).

---

## 🚀 My Goals

### Level 1: Minimal Viable World
- **Goal**: Kick off with a simple grid and random regions.
- **Features**:
  - CLI commands: `move <direction>`, `look`, `quit`.
  - Coordinator tracks position in memory, fakes pod spawning.
  - Region pods dish out random terrain (e.g., `"forest"`).
- **Status**: ✅ Done, evolved into Level 2.
- **Tech**: Go, basic Kubernetes (manual pods), no persistence.

### Level 2: Persistent Grid
- **Goal**: Make the world stick around with auto-managed pods.
- **Features**:
  - Redis stores your position and region types.
  - Coordinator spawns pods via K8s API, cleans up old ones.
  - Supports negative coords (e.g., `-1,1` → `n1,1` labels).
- **Status**: ✅ Done (Feb 28, 2025), live on OKE.
- **Tech**: Go, Kubernetes (OKE), Redis, Docker.
- **Polish Needed**: Smooth out occasional hangs (pod readiness lag).

### Level 3: Dynamic World
- **Goal**: Spice up regions with depth and scale.
- **Features**:
  - Rich terrain (e.g., `"forest with a river"`) via procedural rules.
  - Autoscaling with Horizontal Pod Autoscaler (HPA).
  - Sync borders between pods (e.g., rivers flow across regions).
- **Status**: ⏳ Not started.
- **Tech**: Enhanced Go generation, K8s HPA.

### Level 4: Interactive Ecosystem
- **Goal**: Go big with multiplayer and interactivity.
- **Features**:
  - Multiple explorers sharing the world (pod reuse).
  - NPCs and items (e.g., `"a bandit attacks!"`).
  - Swap CLI for a slick web UI.
- **Status**: ⏳ Not started.
- **Tech**: Go concurrency, K8s multi-pod logic, web frontend.
