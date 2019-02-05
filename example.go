package main

import "flag"

var (
	// vertexColors should be true if the Model you pass in has vertex colors set AND you want the glTF model to use vertex
	// colors. if vertexColors is false, a bitmapped texture atlas will be created from the Materials used in the Model,
	// and the UV coordinates for each vertex will be set to the appropriate pixel on the texture atlas.
	vertexColors = flag.Bool("vc", false, "use vertex colors for mesh colors")

	// if true, a self-contained embedded .gltf file will be generated instead of a self-contained binary .glb.
	embeddedGltf = flag.Bool("e", false, "create embedded .gltf rather than binary .glb model")
)

func main() {
	flag.Parse()

	// set up the single material we'll use.
	plainMaterial := Material{
		DiffuseColor: [3]float32{1.0, 0.0, 0.0},
		Opacity:      1.0,
	}

	// set up a vertex color in case vertex colors are chosen.
	redColor := Vector4{
		R: 1.0,
		G: 0.0,
		B: 0.0,
		A: 1.0,
	}

	// set up the geometry we're going to render:
	vert1 := Vector3{
		X: 0.0,
		Y: 0.0,
		Z: 0.0,
	}

	vert2 := Vector3{
		X: 1.0,
		Y: 0.0,
		Z: 0.0,
	}

	vert3 := Vector3{
		X: 0.0,
		Y: 1.0,
		Z: 0.0,
	}

	normal := Vector3{
		X: 0.0,
		Y: 0.0,
		Z: 1.0,
	}

	// create a mesh using the geometry we just specified
	meshes := Model{
		Meshes: []Geometry{
			Geometry{
				Vertices: []Vertex{
					Vertex{
						Position: vert1,
						Normal:   normal,
						Color:    redColor,
					},
					Vertex{
						Position: vert2,
						Normal:   normal,
						Color:    redColor,
					},
					Vertex{
						Position: vert3,
						Normal:   normal,
						Color:    redColor,
					},
				},
				Faces: []Triangle{
					Triangle{
						TriangleIndices: [3]int32{0, 1, 2},
					},
				},
				Material: plainMaterial,
			},
		},
	}

	// if vertexColors is true, textureAtlas will just be an emtpy bytes.Buffer.
	model, textureAtlas := optimizeModel(meshes, *vertexColors)

	writeGltf(model, textureAtlas, "sample", *embeddedGltf, *vertexColors)
}
