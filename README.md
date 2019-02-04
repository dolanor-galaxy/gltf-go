# gltf-go

This is an incomplete glTF 2.0 serialization library for Go.  It has all of the features that I needed when I wrote it.  It may or may not have the features that you need.  

See `example.go` for an .. example.  (Too 'on the nose'?)  Most/All of the functions & methods in `example.go` should probably go in `GlTFTypes.go` and that file should probably be renamed.

I don't know how to properly set up Go packages for consumption by other code, yet.  I'll get that set up as soon as I have time. 

The most useful piece of code for passers by is likely to be the binary glTF serialization stuff.  That starts on (or around) like 126 of `example.go`.

This "library" is not properly tidied up so don't include it in any projects, yet.  Feel free to copy this code and use it directly in your own program, though.

## Usage:

`go build`

Then, running `gltf-go` will generate example.glb.  Examine in your favorite viewer or validator.