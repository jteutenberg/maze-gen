# Maze Generator

Typical depth-first and spanning tree generators make mazes like:

```
    v
╥╔══╬╗╔══╗
╚╣╔╗║║╚═╗╨
╔╝║╚╝╠═╗╚╗
╚╗╚╗╞╩╗╚═╣
╔╩╡╚═╗║╔╡║
║╞╗╔═╝║╚═╝
╚╗║╚╗╔╩═╗╥
╔╝╠╗║║╔╡╠╝
╚═╝║║╨╚═╩╗
╞══╝╚╗╞══╝
     v
```
with no loops -- only dead ends. Only one decision on the correct solution path.
When starting from the exit, a 22-length path with no decisions at all, and then sight of the entrance.

whereas this project generates:

```
    v  ^
╔╗╔╗║╔╗╚╗╥
║║║╠╝║║╥║║
║╠╣╚═╝║╠╝║
╠╝╨╔══╝╚╗║
║╔═╝╔╗╔╗╚╣
╚╩══╫╝║╚═╝
╞╦╦╡╚╦╫╗╞╗
╔╝╚══╝║╚╗║
╠══╗╔╗╠╗╚╣
╚═╡╚╝╚╝╚═╝
```

though this is 102 tiles (with 2 bridges) instead of 100. Mixture of loops and dead ends, with a preference for loops. 6 decisions on the correct solution path. Only 5 tiles from exit/entrance to a decision point, and not in line of sight.


The solutions can also be generated with internal entrance/exit for stacking levels.
