# gltf-go

This is an incomplete glTF 2.0 serialization library for Go.  It has all of the features that I needed when I wrote it.  It may or may not have the features that you need.  This currently supports static meshes and basic materials and little else.

See `example.go` for an .. example.  (Too 'on the nose'?)  

I don't know how to properly set up Go packages for consumption by other code, yet.  I'll get that set up as soon as I have time. 

This "library" is not properly tidied up so don't include it in any projects, yet.  Feel free to copy this code and use it directly in your own program, though.

## Usage:

`go build`

Then, running `gltf-go` will generate sample.glb.  Examine in your favorite viewer or validator.