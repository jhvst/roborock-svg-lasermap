
# roborock-svg-lasermap

This project generates SVG maps from stationary points using the Xiaomi Roborock V2 (though I don't see why it wouldn't work with the earlier version) vacuum. This project also contains a "classifier", which detects moving objects while the vacuum is running.

For background, please see my Twitter thread: https://twitter.com/toldjuuso/status/1005237185333940224

You can find a demo on YouTube: https://youtu.be/mlAEZZ5fPbo

A visualization of the map generated can be found on https://www.juusohaavisto.com/castle/viz.html. You can copy the source there to create your own visualization. The required file for this is generated on step 2. of the CMD steps.

The map can then be used for stuff like a simple AR PoC: https://youtu.be/OTlOULeNUUo

## Overview steps

Essentially, the way to achieve this is the following:

1. Root ("hack") your Roborock with https://github.com/dgiese/dustcloud
2. Configure Player-service to dispatch LIDAR data to a desktop computer (open firewall ports on the Roborock)
3. Clear the room of clutter
4. Use LIDAR points to create a floorplan of the room, save it (look at /castle/main.go)
5. Use the saved floorplan to only calculate the difference in objects within the room (look at /castle/main.go again)
6. Motion detection achieved (look at /classifier/main.go and /tail/main.go)

## CMD steps

On command line this goes pretty much as following:

1. `ssh user@roborock playerprint -r 0 laser > floorplan.txt`
2. `cat floorplan.txt | go run ./floorplan/main.go`
3. The above command will create you a floorplan with a name `floorplan.csv`. You can render this with the R script given in `/classifier/main.r`. You should also go through this file manually and reduce points which are basically next to each other. The less points, the better. Once done, place the cleaned `floorplan.csv` to `/classifier` folder.
4. Run `watch -n 0.1 Rscript main.r` on the classifier folder. This will start producing file called `Rplots.pdf` each 100ms. Open this with a PDF viewer which autorefreshes on changes.
5. `ssh user@roborock playerprint -r 0 laser | ./castle/castle | ./classifier/classifier > ./classifier/stream.csv`. Keep watching the PDF file for movement detection.
6. Finally, there is a script used in the YouTube demo at `/tail/main.go` which goes through the `stream.csv` file produced on step five whenever it receives a HTTP request to `localhost:8080/tail`. This is used as an endpoint in the demo with Alexa. Essentially, create a new skill on Alexa Developer portal, and when asked about "people in the room" or whatever, make it do a HTTP request.
