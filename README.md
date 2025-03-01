```
       _____________________
      |                     |
      |     DriftScape      |
      |_____________________|

       A Procedural Exploration Game
       Built with Go and Kubernetes

       [ Status: Active ] [ Go: 1.21 ]
       [ K8s: OKE ] [ Redis: Enabled ]


+---------------------[ Overview ]---------------------+
|                                                      |
| DriftScape is an exploration game where you navigate |
| a grid-based world that expands as you move.         |
| Starting at (0,0), you enter commands like "move     |
| north" to generate new regions—such as forests or    |
| plains—each running as a Kubernetes pod. Go handles  |
| the fast, procedural generation, a Coordinator       |
| manages your position, and Redis ensures the world   |
| persists across sessions. Deployed on Oracle         |
| Kubernetes Engine (OKE), it combines technology with |
| a simple gameplay experience.                        |
|                                                      |
+------------------------------------------------------+

+--------------------[ Project Goals ]------------------+
|                                                      |
| >>> Level 1: Minimal Viable World                    |
|                                                      |
| Goal:                                                |
| Establish a basic grid with region generation.       |
|                                                      |
| Features:                                            |
| - CLI supporting "move <direction>", "look", and     |
|   "quit".                                            |
| - Coordinator tracks position in memory,             |
|   simulates pod creation.                            |
| - Region pods generate random terrain types          |
|   (e.g., "forest").                                  |
|                                                      |
| Status:                                              |
| Completed, transitioned to Level 2.                  |
|                                                      |
| Tech:                                                |
| Go, basic Kubernetes with manual pod management,     |
| HTTP/REST.                                           |
|                                                      |
| >>> Level 2: Persistent Grid                         |
|                                                      |
| Goal:                                                |
| Enable persistence and automate pod lifecycle.       |
|                                                      |
| Features:                                            |
| - Redis stores position and region data.             |
| - Coordinator uses the Kubernetes API to spawn and   |
|   remove pods.                                       |
| - Handles negative coordinates (e.g., "-1,1" mapped  |
|   to "n1,1" labels).                                 |
|                                                      |
| Status:                                              |
| Completed on February 28, 2025, running on OKE.      |
|                                                      |
| Tech:                                                |
| Go, Kubernetes (OKE), Redis, Docker, HTTP/REST.      |
|                                                      |
| Bug fixes:                                           |
| Addressed long delays when moving to new regions.    |
|                                                      |
| >>> Level 3: Dynamic World                           |
|                                                      |
| Goal:                                                |
| Enhance regions and improve scalability.             |
|                                                      |
| Features:                                            |
| - Generate detailed terrain (e.g., "forest with a    |
|   river") using procedural methods.                  |
| - Implement autoscaling with Horizontal Pod          |
|   Autoscaler (HPA).                                  |
| - Coordinate region borders (e.g., continuous rivers)|
| - Adopt gRPC for efficient pod communication.        |
|                                                      |
| Status:                                              |
| Not yet started.                                     |
|                                                      |
| Tech:                                                |
| Advanced Go generation, Kubernetes HPA,              |
| gRPC replacing HTTP.                                 |
|                                                      |
| >>> Level 4: Interactive Ecosystem                   |
|                                                      |
| Goal:                                                |
| Support multiple players and interactive elements.   |
|                                                      |
| Features:                                            |
| - Allow multiple users to share the same world with  |
|   pod reuse.                                         |
| - Introduce NPCs and items (e.g., "a dog appears").  |
| - Replace CLI with a web-based interface.            |
|                                                      |
| Status:                                              |
| Not yet started.                                     |
|                                                      |
| Tech:                                                |
| Go concurrency, Kubernetes multi-pod management,     |
| gRPC, web frontend.                                  |
|                                                      |
+------------------------------------------------------+
```
