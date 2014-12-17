# routeplanner

A simple route planning software with a focus on bicycles written in Go (using OpenStreetMap data).

Just putting the source code online for those of you who are interested in the basic techniques and reading the code. I developed this route planner mainly for learning purposes. I never planned to release the code which is why most of the code comments are in German (sorry). Hence my approach is far from complete or ready for being used in production; it's likely that this code won't run on your computer because of invalid paths etc (I have to clean up the code first). I'm currently working on a new version where I try to take all leasons learned into account.

I put up a **demo page** for the Berlin/Germany area: 
**[http://route.florian-schlachter.de](http://route.florian-schlachter.de)** ([API available](http://route.florian-schlachter.de/api))

Anyway, the route planner works quite nice even given its early stage. :-)

## Features

 * Route calculation with profile support (for now, *cars* and *bicycles*)
 * JSON-API for route calculation and point discovery
 * Calculation of ETA, duration and route's bike compatiblity (in percent)
 * General considerations for routes:
    * Speed limitations
    * Turn constraints
    * Driving directions
    * Traffic signals
    * Way types (private/public accesses)
    * Indication for tunnel entrances/exits
 * Bicycle profile:
    * Street conditions
    * Shortcuts using footways
    * Lightning conditions
    * Prefers "nice" ways (i.e. green and calm areas)
    * Level crossings
    * Barriers
 * Car profile:
    * Support for motorways, route indication for entrace/exit ramps

## Blog posts

The following blog posts are written in German, sorry folks.

 * [Routenplaner für Fahrradtouren in Golang](https://www.florian-schlachter.de/post/routenplaner/) [May 15th 2014]
 * [Routenplaner in Alpha-Version online](https://www.florian-schlachter.de/post/routenplaner-online/) [May 29th 2014]
 * [Ausflugsziele/Touren in Berlin und Umland mit dem Fahrrad](https://www.florian-schlachter.de/post/routenplaner-ausflugsziele/) [Jun 1st 2014]
