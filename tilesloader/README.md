# tilesloader

`/load-tiles`
Accepts mapget.MapDescription (JSON) in request body.
MapDescription is used to specify which tiles to download
(and from where).

`/load-tiles/save`
Download tiles and save them to the tile database.
Pass them to further handling after.

`/load-tiles/handle`
Download tiles without saving them to the tile database.
Pass them to further handling after.
