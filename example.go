package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"strings"
)

// vertexColors should be true if the Model you pass in has vertex colors set AND you want the glTF model to use vertex
// colors. if vertexColors is false, a bitmapped texture atlas will be created from the Materials used in the Model,
// and the UV coordinates for each vertex will be set to the appropriate pixel on the texture atlas.
var vertexColors = true

// if true, a self-contained embedded .gltf file will be generated instead of a self-contained binary .glb.
var embeddedGltf = false

func main() {
	// set up the single material we'll use.
	plain := Material{
		AmbientColor: [3]float32{1.0, 0.0, 0.0},
		Opacity:      1.0,
	}

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
				Material: plain,
			},
		},
	}

	model, atlas := optimizeModel(meshes)

	writeGltf(model, atlas, "example", embeddedGltf)
}

func writeGltf(model Model, atlas bytes.Buffer, filename string, embeddedGltf bool) {
	gltfDoc := ToGltfDoc(model, atlas)

	gltfDoc.Meshes[0].Name = filename
	gltfDoc.Nodes[0].Name = filename

	gltfFileContents := []byte{}
	gltfOutputFile := ""

	if embeddedGltf {
		gltfFileContents = SerializeEmbeddedGlTF(gltfDoc)
		gltfOutputFile = filename + ".gltf"
	} else {
		gltfFileContents = SerializeBinaryGlTF(gltfDoc)
		gltfOutputFile = filename + ".glb"
	}

	gltfOutput, _ := os.Create(gltfOutputFile)
	gltfWriter := bufio.NewWriter(gltfOutput)

	gltfWriter.Write(gltfFileContents)
	gltfWriter.Flush()
}

// SerializeBinaryGlTF renders a GlTF document to a byte slice containing a binary glTF document.
func SerializeBinaryGlTF(gltfDoc GlTF) []byte {
	outBuf := bytes.NewBuffer(gltfDoc.Buffers[0].Bytes)

	// get the JSON content for the binary file.
	outJSON, _ := json.Marshal(gltfDoc)

	// get the JSON size, and the number of padding spaces required.
	outJSONSize := uint32(len(outJSON))

	// calculate the number of spaces required to make the size of the JSON modulus 4 equal to 0.
	outJSONPaddingNeeded := (4 - (outJSONSize & 3)) & 3
	outJSONSize += outJSONPaddingNeeded

	// get the binary mesh data buffer length and the number of padding nulls required.
	outBufSize := uint32(outBuf.Len())

	// calculate the number of padding null bytes that makes the size of the binary buffer modulus 4 equal to 0.
	outBufNullsNeeded := (4 - (outBufSize & 3)) & 3
	outBufSize += outBufNullsNeeded

	// calculate the final GLB file size
	glbSize := uint32(0)   // entire file length =
	glbSize += 4           // gltf magic number field length +
	glbSize += 4           // version number field length +
	glbSize += 4           // total file size field length +
	glbSize += 4           // json chunk type header length +
	glbSize += 4           // json chunk declaration length +
	glbSize += outJSONSize // json payload length +
	glbSize += 4           // binary chunk type header length +
	glbSize += 4           // binary chunk declaration length +
	glbSize += outBufSize  // binary payload length.

	// set up the output byte array
	outData := new(bytes.Buffer)

	// write the magic number
	outData.WriteString("glTF")

	// write the glTF version
	binary.Write(outData, binary.LittleEndian, uint32(2))

	// write the whole filesize
	binary.Write(outData, binary.LittleEndian, glbSize)

	// write the JSON chunk size
	binary.Write(outData, binary.LittleEndian, outJSONSize)

	// write the JSON chunk header
	outData.WriteString("JSON")

	// write the JSON chunk itself
	outData.WriteString(string(outJSON))

	// pad the JSON with spaces, if required.
	outData.WriteString(strings.Repeat(" ", int(outJSONPaddingNeeded)))

	// write the binary chunk length
	binary.Write(outData, binary.LittleEndian, outBufSize)

	// write the binary chunk type
	outData.WriteString("BIN\x00")

	// write the binary chunk itself
	outBuf.WriteTo(outData)

	// pad with nulls if required
	for outBufNullsNeeded > 0 {
		outData.WriteByte(byte(0))
		outBufNullsNeeded--
	}

	// done.
	return outData.Bytes()
}

// SerializeEmbeddedGlTF renders a GlTF document to a byte slice containing an embedded glTF document.
func SerializeEmbeddedGlTF(gltfDoc GlTF) []byte {
	outBuf := bytes.NewBuffer(gltfDoc.Buffers[0].Bytes)

	// ASCII glTF is easier for the developer of this application.
	gltfDoc.Buffers[0].URI = "data:application/gltf-buffer;base64," + base64.StdEncoding.EncodeToString(outBuf.Bytes())

	outData, err := json.MarshalIndent(gltfDoc, "", "    ")

	printIf(err != nil, "couldn't marshal json.")

	return outData
}

// Appends the supplied triangle indices to the supplied bytes.Buffer.  It is up to the calling function to observe the
// length of the buffer before and after this function is called. Returns the minimum and maximum values observed in the
// supplied indices so they can be defined in the glTF file that uses the appended data.
func addTriangleArrayToBuffer(outBuf *bytes.Buffer, indices []Triangle) (min, max uint32) {
	min = uint32(100)
	max = uint32(0)

	for _, i := range indices {
		binary.Write(outBuf, binary.LittleEndian, i.TriangleIndices[0])
		binary.Write(outBuf, binary.LittleEndian, i.TriangleIndices[1])
		binary.Write(outBuf, binary.LittleEndian, i.TriangleIndices[2])

		max = uint32(math.Max(float64(max), float64(i.TriangleIndices[0])))
		max = uint32(math.Max(float64(max), float64(i.TriangleIndices[1])))
		max = uint32(math.Max(float64(max), float64(i.TriangleIndices[2])))

		min = uint32(math.Min(float64(min), float64(i.TriangleIndices[0])))
		min = uint32(math.Min(float64(min), float64(i.TriangleIndices[1])))
		min = uint32(math.Min(float64(min), float64(i.TriangleIndices[2])))
	}

	return min, max
}

// see the Vector3 version.
func addVector2ArrayToBuffer(outBuf *bytes.Buffer, data *[]Vector2) (min, max Vector2) {
	highestU := float32(-100)
	highestV := float32(-100)

	lowestU := float32(100)
	lowestV := float32(100)

	for _, v := range *data {
		binary.Write(outBuf, binary.LittleEndian, v.U)
		binary.Write(outBuf, binary.LittleEndian, v.V)

		highestU = float32(math.Max(float64(highestU), float64(v.U)))
		highestV = float32(math.Max(float64(highestV), float64(v.V)))

		lowestU = float32(math.Min(float64(lowestU), float64(v.U)))
		lowestV = float32(math.Min(float64(lowestV), float64(v.V)))
	}

	min = Vector2{U: lowestU, V: lowestV}
	max = Vector2{U: highestU, V: highestV}

	return min, max
}

// Appends the supplied Vector3 slice to the supplied bytes.Buffer.  It is up to the calling function to observe the
// length of the buffer before and after modification so that the appropriate BufferView and Accessor can be created.
// Returns new Vector3s containing the minimum and maximum observed values for each component in the supplied Vector3
// slice so that they can be used in the glTF file that uses the appended data.
func addVector3ArrayToBuffer(outBuf *bytes.Buffer, data *[]Vector3) (min, max Vector3) {
	highestX := float32(-100)
	highestY := float32(-100)
	highestZ := float32(-100)

	lowestX := float32(100)
	lowestY := float32(100)
	lowestZ := float32(100)

	for _, v := range *data {
		binary.Write(outBuf, binary.LittleEndian, v.X)
		binary.Write(outBuf, binary.LittleEndian, v.Y)
		binary.Write(outBuf, binary.LittleEndian, v.Z)

		highestX = float32(math.Max(float64(highestX), float64(v.X)))
		highestY = float32(math.Max(float64(highestY), float64(v.Y)))
		highestZ = float32(math.Max(float64(highestZ), float64(v.Z)))

		lowestX = float32(math.Min(float64(lowestX), float64(v.X)))
		lowestY = float32(math.Min(float64(lowestY), float64(v.Y)))
		lowestZ = float32(math.Min(float64(lowestZ), float64(v.Z)))
	}

	min = Vector3{X: lowestX, Y: lowestY, Z: lowestZ}
	max = Vector3{X: highestX, Y: highestY, Z: highestZ}

	return min, max
}

func addVector4ArrayToBuffer(outBuf *bytes.Buffer, data *[]Vector4) (min, max Vector4) {
	highestR := float32(-100)
	highestG := float32(-100)
	highestB := float32(-100)
	highestA := float32(-100)

	lowestR := float32(100)
	lowestG := float32(100)
	lowestB := float32(100)
	lowestA := float32(100)

	for _, v := range *data {
		binary.Write(outBuf, binary.LittleEndian, v.R)
		binary.Write(outBuf, binary.LittleEndian, v.G)
		binary.Write(outBuf, binary.LittleEndian, v.B)
		binary.Write(outBuf, binary.LittleEndian, v.A)

		highestR = float32(math.Max(float64(highestR), float64(v.R)))
		highestG = float32(math.Max(float64(highestG), float64(v.G)))
		highestB = float32(math.Max(float64(highestB), float64(v.B)))
		highestA = float32(math.Max(float64(highestA), float64(v.A)))

		lowestR = float32(math.Min(float64(lowestR), float64(v.R)))
		lowestG = float32(math.Min(float64(lowestG), float64(v.G)))
		lowestB = float32(math.Min(float64(lowestB), float64(v.B)))
		lowestA = float32(math.Min(float64(lowestA), float64(v.A)))
	}

	min = Vector4{R: lowestR, G: lowestG, B: lowestB, A: lowestA}
	max = Vector4{R: highestR, G: highestG, B: highestB, A: highestA}

	return min, max
}

func getAccessorIndexFromVector2(outBuf *bytes.Buffer, vectors []Vector2, gltfBufferViews *[]BufferView, gltfAccessors *[]Accessor) (accessorIndex int) {
	byteOffset := outBuf.Len()
	min, max := addVector2ArrayToBuffer(outBuf, &vectors)
	byteLength := outBuf.Len() - byteOffset

	verticesBufferView := BufferView{
		Buffer:     0,
		ByteOffset: byteOffset,
		ByteLength: byteLength,
		ByteStride: 8,
		Target:     34962,
	}

	*gltfBufferViews = append(*gltfBufferViews, verticesBufferView)

	verticesAccessor := Accessor{
		BufferView:    len(*gltfBufferViews) - 1,
		ByteOffset:    0,
		ComponentType: 5126,
		Count:         len(vectors),
		Type:          "VEC2",
		Max:           []float32{max.U, max.V},
		Min:           []float32{min.U, min.V},
	}

	*gltfAccessors = append(*gltfAccessors, verticesAccessor)

	return len(*gltfAccessors) - 1
}

// Appends an array of Vector3 to a bytes.Buffer, then generates and adds the appropriate glTF BufferView and glTF
// accessor to the supplied slices, then returns those new, modified slices.
func getAccessorIndexFromVector3(outBuf *bytes.Buffer, vectors []Vector3, gltfBufferViews *[]BufferView, gltfAccessors *[]Accessor) (accessorIndex int) {
	byteOffset := outBuf.Len()
	min, max := addVector3ArrayToBuffer(outBuf, &vectors)
	byteLength := outBuf.Len() - byteOffset

	verticesBufferView := BufferView{
		Buffer:     0,
		ByteOffset: byteOffset,
		ByteLength: byteLength,
		ByteStride: 12,
		Target:     34962,
	}

	*gltfBufferViews = append(*gltfBufferViews, verticesBufferView)

	verticesAccessor := Accessor{
		BufferView:    len(*gltfBufferViews) - 1,
		ByteOffset:    0,
		ComponentType: 5126,
		Count:         len(vectors),
		Type:          "VEC3",
		Max:           []float32{max.X, max.Y, max.Z},
		Min:           []float32{min.X, min.Y, min.Z},
	}

	*gltfAccessors = append(*gltfAccessors, verticesAccessor)

	return len(*gltfAccessors) - 1
}

func getAccessorIndexFromVector4(outBuf *bytes.Buffer, vectors []Vector4, gltfBufferViews *[]BufferView, gltfAccessors *[]Accessor) (accessorIndex int) {
	byteOffset := outBuf.Len()
	min, max := addVector4ArrayToBuffer(outBuf, &vectors)
	byteLength := outBuf.Len() - byteOffset

	verticesBufferView := BufferView{
		Buffer:     0,
		ByteOffset: byteOffset,
		ByteLength: byteLength,
		ByteStride: 16,
		Target:     34962,
	}

	*gltfBufferViews = append(*gltfBufferViews, verticesBufferView)

	verticesAccessor := Accessor{
		BufferView:    len(*gltfBufferViews) - 1,
		ByteOffset:    0,
		ComponentType: 5126,
		Count:         len(vectors),
		Type:          "VEC4",
		Max:           []float32{max.R, max.G, max.B, max.A},
		Min:           []float32{min.R, min.G, min.B, min.A},
	}

	*gltfAccessors = append(*gltfAccessors, verticesAccessor)

	return len(*gltfAccessors) - 1
}

// Appends an array of triangle indices to the supplied bytes.Buffer, then generates and adds the appropriate glTF
// BufferView and glTF Accessor to the supplied slices, then returns the new, modified slices.
func getAccessorIndexFromIndices(outBuf *bytes.Buffer, indices []Triangle, gltfBufferViews *[]BufferView, gltfAccessors *[]Accessor) (accessorIndex int) {
	byteOffset := outBuf.Len()
	min, max := addTriangleArrayToBuffer(outBuf, indices)
	byteLength := outBuf.Len() - byteOffset

	indicesBufferView := BufferView{
		Buffer:     0,
		ByteOffset: byteOffset,
		ByteLength: byteLength,
		Target:     34963,
	}

	*gltfBufferViews = append(*gltfBufferViews, indicesBufferView)

	indicesAccessor := Accessor{
		BufferView:    len(*gltfBufferViews) - 1,
		ByteOffset:    0,
		ComponentType: 5125,
		Count:         len(indices) * 3,
		Type:          "SCALAR",
		Max:           []float32{float32(max)},
		Min:           []float32{float32(min)},
	}

	*gltfAccessors = append(*gltfAccessors, indicesAccessor)

	return len(*gltfAccessors) - 1
}

// Adds a material to the supplied []GltfMaterial array if it is not already present,
// and returns this material's index in that array.
func addMaterial(material GltfMaterial, gltfMaterials []GltfMaterial) (materialIndex int, newMaterials []GltfMaterial) {
	// if material exists, return its index.
	for i, m := range gltfMaterials {
		if areMaterialsEqual(m, material) {
			return i, gltfMaterials
		}
	}

	// if material does not exist in gltfMaterials, add the material to the collection, and ...
	newMaterials = append(gltfMaterials, material)

	// ...return the new index.
	return len(newMaterials) - 1, newMaterials
}

// compare materials for equality.  This should probably instead return -1, 0, or 1 so they can be sorted by hue, then
// opacity.  I don't know if I would ever need to sort materials, though.
func areMaterialsEqual(a GltfMaterial, b GltfMaterial) bool {
	sameR := a.PbrMetallicRoughness.BaseColorFactor[0] == b.PbrMetallicRoughness.BaseColorFactor[0]
	sameG := a.PbrMetallicRoughness.BaseColorFactor[1] == b.PbrMetallicRoughness.BaseColorFactor[1]
	sameB := a.PbrMetallicRoughness.BaseColorFactor[2] == b.PbrMetallicRoughness.BaseColorFactor[2]
	sameA := a.PbrMetallicRoughness.BaseColorFactor[3] == b.PbrMetallicRoughness.BaseColorFactor[3]

	sameMe := a.PbrMetallicRoughness.MetallicFactor == b.PbrMetallicRoughness.MetallicFactor
	sameRo := a.PbrMetallicRoughness.RoughnessFactor == b.PbrMetallicRoughness.RoughnessFactor

	return sameR && sameG && sameB && sameA && sameMe && sameRo
}

// TODO: support more material and appearance features, despite their apparent lack of use by our models.
// This creates a new gltfMaterial based on the properties in the supplied Material.
func gltfMaterial(material Material) GltfMaterial {
	outMaterial := GltfMaterial{
		DoubleSided: true,
		PbrMetallicRoughness: MaterialPbrMetallicRoughness{
			BaseColorFactor: []float64{
				float64(material.DiffuseColor[0]),
				float64(material.DiffuseColor[1]),
				float64(material.DiffuseColor[2]),
				float64(material.Opacity),
			},
			MetallicFactor:  0.0,
			RoughnessFactor: 1.0 - (float64(material.SpecularPower) / 128.0),
		},
	}

	if material.Opacity < 1 {
		outMaterial.AlphaMode = "BLEND"
	}

	return outMaterial
}

// TODO: rename this to 'applyMaterialStrategy' probably since that's what it does.
func optimizeModel(meshes Model) (Model, bytes.Buffer) {
	// set up the fully merged Geometry structure.
	finalVertices := []Vertex{}
	finalFaces := []Triangle{}
	imageData := new(bytes.Buffer)

	if !vertexColors {
		// the texture atlas case.

		// set up the texture atlas and populate it as you go through the Geometry objects.
		img := image.NewRGBA(image.Rect(0, 0, 32, 32))

		for i, mesh := range meshes.Meshes {
			vertexOffset := int32(len(finalVertices))

			//* color correction: scale all colors from 0-1 to 0.04-0.85 because gltf uses Physically Based Rendering.
			//* https://seblagarde.wordpress.com/2011/08/17/feeding-a-physical-based-lighting-mode/
			// TODO: scale all colors by the same amount; just enough to bring the brightest and darkest colors into range.
			dR := mapRange(float64(mesh.Material.DiffuseColor[0]), 0.0, 1.0, 0.04, 0.85)
			dG := mapRange(float64(mesh.Material.DiffuseColor[1]), 0.0, 1.0, 0.04, 0.85)
			dB := mapRange(float64(mesh.Material.DiffuseColor[2]), 0.0, 1.0, 0.04, 0.85)

			r := uint8(dR * 255)
			g := uint8(dG * 255)
			b := uint8(dB * 255)
			a := uint8(mesh.Material.Opacity * 255)

			x := i % 32
			y := i / 32

			// set the pixel on the texture atlas
			color := color.RGBA{r, g, b, a}
			img.Set(x, y, color)

			// add a reference to this pixel for all the vertices that use this color.
			for _, vertex := range mesh.Vertices {
				vertex.UV = Vector2{
					U: (float32(x) / 32.0) + (0.5 / 32.0),
					V: (float32(y) / 32.0) + (0.5 / 32.0),
				}

				finalVertices = append(finalVertices, vertex)
			}

			// add the triangles to the new monolithic mesh, using the new indices.
			for _, triangle := range mesh.Faces {
				f := Triangle{
					TriangleIndices: [3]int32{
						triangle.TriangleIndices[0] + vertexOffset,
						triangle.TriangleIndices[1] + vertexOffset,
						triangle.TriangleIndices[2] + vertexOffset,
					},
				}

				finalFaces = append(finalFaces, f)
			}
		}

		// finally, save out the texture atlas.  we're saving to a bytes.Buffer in this case, not a file.
		png.Encode(imageData, img)
	} else {
		// The vertex color case.
		for _, mesh := range meshes.Meshes {
			vertexOffset := int32(len(finalVertices))

			for _, vertex := range mesh.Vertices {
				vertex.Color.R = float32(mapRange(float64(vertex.Color.R), 0.0, 1.0, 0.04, 0.85))
				vertex.Color.G = float32(mapRange(float64(vertex.Color.G), 0.0, 1.0, 0.04, 0.85))
				vertex.Color.B = float32(mapRange(float64(vertex.Color.B), 0.0, 1.0, 0.04, 0.85))

				finalVertices = append(finalVertices, vertex)
			}

			for _, triangle := range mesh.Faces {
				f := Triangle{
					TriangleIndices: [3]int32{
						triangle.TriangleIndices[0] + vertexOffset,
						triangle.TriangleIndices[1] + vertexOffset,
						triangle.TriangleIndices[2] + vertexOffset,
					},
				}

				finalFaces = append(finalFaces, f)
			}
		}
	}

	// create a new Model using everything we just created above.
	meshes = Model{
		Meshes: []Geometry{
			Geometry{
				Vertices: finalVertices,
				Faces:    finalFaces,
				Material: Material{
					AmbientColor:  [3]float32{1.0, 1.0, 1.0},
					DiffuseColor:  [3]float32{1.0, 1.0, 1.0},
					SpecularColor: [3]float32{1.0, 1.0, 1.0},
					SpecularPower: 128,
					EmissiveColor: [3]float32{1.0, 1.0, 1.0},
					Opacity:       1.0,
				},
			},
		},
	}

	// return it.
	return meshes, *imageData
}

// ToGltfDoc converts a model to a GlTF object, ready for serialization.
func ToGltfDoc(model Model, atlas bytes.Buffer) GlTF {
	gltfBufferViews := []BufferView{}
	gltfAccessors := []Accessor{}
	gltfBuffers := []GltfBuffer{}
	gltfMeshes := []Mesh{}
	gltfNodes := []Node{}
	gltfScenes := []Scene{}
	gltfMaterials := []GltfMaterial{}

	outBuf := new(bytes.Buffer)

	associations := []meshInfoAssociation{}

	for _, mesh := range model.Meshes {
		thisMaterial := gltfMaterial(mesh.Material)

		uvAccessorIndex := -1
		vertexColorAccessorIndex := -1

		meshIndicesAccessorIndex := getAccessorIndexFromIndices(outBuf, mesh.Faces, &gltfBufferViews, &gltfAccessors)
		meshVertexAccessorIndex := getAccessorIndexFromVector3(outBuf, getVertices(mesh), &gltfBufferViews, &gltfAccessors)
		meshNormalAccessorIndex := getAccessorIndexFromVector3(outBuf, getNormals(mesh), &gltfBufferViews, &gltfAccessors)

		if !vertexColors {
			uvAccessorIndex = getAccessorIndexFromVector2(outBuf, getUVCoords(mesh), &gltfBufferViews, &gltfAccessors)
			baseColorTexture := make(map[string]int)
			baseColorTexture["index"] = 0

			thisMaterial.PbrMetallicRoughness.BaseColorTexture = baseColorTexture
		} else {
			vertexColorAccessorIndex = getAccessorIndexFromVector4(outBuf, getVertexColors(mesh), &gltfBufferViews, &gltfAccessors)

			thisMaterial.PbrMetallicRoughness.BaseColorTexture = nil
		}

		thisMaterial.PbrMetallicRoughness.BaseColorFactor = []float64{1.0, 1.0, 1.0, 1.0}

		materialIndex, newGltfMaterials := addMaterial(thisMaterial, gltfMaterials)

		gltfMaterials = newGltfMaterials

		accessorAssociation := meshInfoAssociation{
			MeshIndicesAccessorIndex:  meshIndicesAccessorIndex,
			MeshMaterialIndex:         materialIndex,
			MeshNormalsAccessorIndex:  meshNormalAccessorIndex,
			MeshVerticesAccessorIndex: meshVertexAccessorIndex,
		}

		if !vertexColors {
			accessorAssociation.MeshUVAccessorIndex = uvAccessorIndex
		} else {
			accessorAssociation.MeshVertexColorAccessorIndex = vertexColorAccessorIndex
		}

		associations = append(associations, accessorAssociation)

		// make sure we're aligned properly.
		if outBuf.Len()%4 != 0 {
			fmt.Println(" this should't happen. ")
		}
	}

	nodeList := []int{}

	meshPrimitives := []MeshPrimitive{}

	for _, assoc := range associations {
		meshPrimitiveAttributes := make(map[string]int)
		meshPrimitiveAttributes["POSITION"] = assoc.MeshVerticesAccessorIndex
		meshPrimitiveAttributes["NORMAL"] = assoc.MeshNormalsAccessorIndex

		if !vertexColors {
			meshPrimitiveAttributes["TEXCOORD_0"] = assoc.MeshUVAccessorIndex
		} else {
			meshPrimitiveAttributes["COLOR_0"] = assoc.MeshVertexColorAccessorIndex
		}

		mp := MeshPrimitive{
			Attributes: meshPrimitiveAttributes,
			Indices:    assoc.MeshIndicesAccessorIndex,
			Material:   assoc.MeshMaterialIndex,
		}

		meshPrimitives = append(meshPrimitives, mp)
	}

	gltfMeshes = append(gltfMeshes, Mesh{Primitives: meshPrimitives})
	gltfNodes = append(gltfNodes, Node{Mesh: len(gltfMeshes) - 1})
	nodeList = append(nodeList, len(gltfNodes)-1)
	scene := Scene{Nodes: nodeList}
	gltfScenes = append(gltfScenes, scene)
	rootSceneIndex := len(gltfScenes) - 1

	gltfBuffer := GltfBuffer{ByteLength: outBuf.Len()}
	gltfBuffer.Bytes = outBuf.Bytes()

	gltfBuffers = append(gltfBuffers, gltfBuffer)

	gltfDoc := GlTF{
		Accessors: gltfAccessors,
		Asset: Asset{
			Version:   "2.0",
			Generator: "gltf-go, https://github.com/naikrovek/gltf-go/", // feel free to change this; i don't mind.
		},
		Buffers:     gltfBuffers,
		BufferViews: gltfBufferViews,
		Materials:   gltfMaterials,
		Meshes:      gltfMeshes,
		Nodes:       gltfNodes,
		Scene:       rootSceneIndex,
		Scenes:      gltfScenes,
	}

	if !vertexColors {
		gltfDoc.Images = []GltfImage{GltfImage{URI: "data:image/png;base64," + base64.StdEncoding.EncodeToString(atlas.Bytes())}}
		gltfDoc.Textures = []GltfTexture{GltfTexture{Source: 0}}
	}

	return gltfDoc
}

func getVertices(mesh Geometry) []Vector3 {
	results := []Vector3{}

	for _, m := range mesh.Vertices {
		results = append(results, m.Position)
	}

	return results
}

func getNormals(mesh Geometry) []Vector3 {
	results := []Vector3{}

	for _, m := range mesh.Vertices {
		results = append(results, m.Normal)
	}

	return results
}

func getUVCoords(mesh Geometry) []Vector2 {
	results := []Vector2{}

	for _, m := range mesh.Vertices {
		results = append(results, m.UV)
	}

	return results
}

func getVertexColors(mesh Geometry) []Vector4 {
	results := []Vector4{}

	for _, m := range mesh.Vertices {
		results = append(results, m.Color)
	}

	return results
}

// Model is a wrapper around []Geometry meshes.
type Model struct {
	//Materials []Material `json:"materials,omitempty"`
	Meshes []Geometry `json:"meshes,omitempty"`
}

// Geometry ...
type Geometry struct {
	Vertices []Vertex   `json:"vertices,omitempty"`
	Faces    []Triangle `json:"faces,omitempty"`
	Material Material   `json:"material"`
}

// Material as defined in the binary file
type Material struct {
	AmbientColor  [3]float32 `json:"ambientColor,omitempty"`
	DiffuseColor  [3]float32 `json:"diffuseColor,omitempty"`
	SpecularColor [3]float32 `json:"specularColor,omitempty"`
	SpecularPower float32    `json:"specularPower"`
	EmissiveColor [3]float32 `json:"emissiveColor,omitempty"`
	Opacity       float32    `json:"opacity"`
}

// Triangle ...
type Triangle struct {
	TriangleIndices [3]int32 `json:"triangle"`
}

// Vector4 is often used for colors or maybe a Vector3 with a magnitude?  I don't know.  I'm going to use it for colors.
type Vector4 struct {
	R float32 `json:"r"`
	G float32 `json:"g"`
	B float32 `json:"b"`
	A float32 `json:"a"`
}

// Vector3 ...
type Vector3 struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

// Vector2 ...
type Vector2 struct {
	U float32 `json:"u"`
	V float32 `json:"v"`
}

// Vertex ...
type Vertex struct {
	Color    Vector4 `json:"color,omitempty"`
	Position Vector3 `json:"position,omitempty"`
	Normal   Vector3 `json:"normal,omitempty"`
	UV       Vector2 `json:"uv,omitempty"`
}

func failIf(condition bool, message ...interface{}) {
	if condition {
		log.Fatal(message...)
	}
}

func printIf(condition bool, message ...interface{}) {
	if condition {
		fmt.Println(message...)
	}
}

func logIf(condition bool, message ...interface{}) {
	if condition {
		log.Println(message...)
	}
}

func mapRange(x float64, inMin float64, inMax float64, outMin float64, outMax float64) float64 {
	return (x-inMin)*(outMax-outMin)/(inMax-inMin) + outMin
}
