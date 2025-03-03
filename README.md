<div style="text-align: center;">
  <h1>DriftScape</h1>
  <div style="display: flex; justify-content: center; align-items: center; flex-direction: column;">
    <img src="https://www.upload.ee/image/17808786/grids.jpeg" height="300" />
    <p>A Procedural Exploration Game<br>Built with Go and Kubernetes</p>
    <p><strong>Status:</strong> Level 3</p>
    <p><strong>Go:</strong> 1.21</p>
    <p><strong>Kubernetes:</strong> OKE</p>
    <p><strong>Redis:</strong> Enabled</p>
  </div>
</div>

## Overview

DriftScape is a procedural exploration game built on a grid system, developed using Go and deployed on Oracle Kubernetes Engine (OKE). Players navigate an expanding world where regions are dynamically generated as Kubernetes pods, with state persistence managed by Redis.

## Progress

The project has progressed through three levels:

### Level 1: Minimal Viable World
- Established a basic grid with random terrain types (e.g., forest, plains), controlled via CLI. Pods were simulated, with no persistence.

### Level 2: Persistent Grid
- Introduced Redis for persistent storage of player position and region data. Automated pod creation and cleanup upon movement.

### Level 3: Dynamic World
- Enhanced regions with detailed terrain (e.g., "forest with a river"), implemented HPA for pod scaling, transitioned to gRPC for efficient pod communication, and added basic border syncing (e.g., rivers connecting south-to-north).

## Next Steps

**Level 4: Interactive Ecosystem**  
Planned enhancements include:

- **Multiple Users:** Enable shared region pods for multiplayer, tracking individual player positions in Redis.
- **NPCs and Items:** Introduce dynamic elements such as non-player characters (e.g., bandits) and items (e.g., swords) within regions.
- **Web Interface:** Replace the CLI with a web-based UI featuring a visual grid and real-time updates via WebSocket.
