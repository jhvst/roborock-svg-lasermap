
d = read.csv("floorplan.csv", header=F)
plot(d)
lines(d)

q = read.csv("stream.csv", header=F)
points(tail(q, 30), pch="*")
