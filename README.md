# gltf-go

This is an incomplete glTF 2.0 serialization library for Go.  It has all of the features that I needed when I wrote it.  It may or may not have the features that you need.  

You feed this a list of vertices and triangles, and by God it'll get you a glTF model.  See example.go for an .. example.  (Too 'on the nose'?)

I don't know how to properly set up Go packages for consumption by other code, yet.  I'll get that set up as soon as I have time. 

The most useful piece of code for passers by is likely to be the binary glTF serialization stuff.  That starts on (or around) like 126 of `example.go`.

This "library" is not properly tidied up so don't include it in any projects, yet.  Feel free to copy this code and use it directly in your own program, though.

## Usage:

`go build`

Then,

running `gltf-go` will generate (and overwrite) example.glb.  Examine in your favorite viewer or validator.

Currently, the texture atlas method doesn't work.  Vertex colors work fine, but you're stuck with RGBA there.