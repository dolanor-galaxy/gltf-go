package main

// Refer to the glTF 2.0 spec (https://github.com/KhronosGroup/glTF/tree/master/specification/2.0) to know what these
// objects are and what they refer to.  If I were to reproduce that info here it would just be a copy & paste job and
// the spec is authoritative.

// Accessor ...
type Accessor struct {
	BufferView    int         `json:"bufferView" validator:"gte=0"`
	ByteOffset    int         `json:"byteOffset" validator:"gte=0"`
	ComponentType interface{} `json:"componentType,omitempty"`
	Count         int         `json:"count" validator:"gte=1"`
	Type          interface{} `json:"type,omitempty"`
	Extensions    interface{} `json:"extensions,omitempty"`
	Extras        interface{} `json:"extras,omitempty"`
	Max           []float32   `json:"max,omitempty"`
	Min           []float32   `json:"min,omitempty"`
	Name          interface{} `json:"name,omitempty"`
	Normalized    bool        `json:"normalized,omitempty"`
	Sparse        interface{} `json:"sparse,omitempty"`
}

// Asset ...
type Asset struct {
	Copyright  string      `json:"copyright,omitempty"`
	Extensions interface{} `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Generator  string      `json:"generator,omitempty"`
	MinVersion string      `json:"minVersion,omitempty"`
	Version    string      `json:"version,omitempty"`
}

// BufferView ...
type BufferView struct {
	Buffer     int         `json:"buffer" validator:"gte=0"`
	ByteLength int         `json:"byteLength" validator:"gte=1"`
	ByteOffset int         `json:"byteOffset" validator:"gte=0"`
	ByteStride int         `json:"byteStride,omitempty" validator:"gte=4, lte=252"`
	Extensions interface{} `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Name       interface{} `json:"name,omitempty"`
	Target     interface{} `json:"target,omitempty"`
}

// GlTF ...
type GlTF struct {
	Accessors          []Accessor     `json:"accessors,omitempty"`
	Asset              interface{}    `json:"asset,omitempty"`
	Buffers            []GltfBuffer   `json:"buffers,omitempty"`
	BufferViews        []BufferView   `json:"bufferViews,omitempty"`
	Extensions         interface{}    `json:"extensions,omitempty"`
	ExtensionsRequired []string       `json:"extensionsRequired,omitempty"`
	ExtensionsUsed     []string       `json:"extensionsUsed,omitempty"`
	Images             []GltfImage    `json:"images,omitempty"`
	Materials          []GltfMaterial `json:"materials,omitempty"`
	Meshes             []Mesh         `json:"meshes,omitempty"`
	Nodes              []Node         `json:"nodes,omitempty"`
	Scene              int            `json:"scene"`
	Scenes             []Scene        `json:"scenes,omitempty"`
	Textures           []GltfTexture  `json:"textures,omitempty"`
}

// GltfTexture ...
type GltfTexture struct {
	Source interface{} `json:"source,omitempty"`
}

// GltfImage ...
type GltfImage struct {
	URI string `json:"uri,omitempty"`
}

// GltfBuffer ...
type GltfBuffer struct {
	ByteLength int         `json:"byteLength" validator:"gte=1"`
	Bytes      []byte      `json:"-"` // don't serialize this, not part of the spec.
	Extensions interface{} `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Name       interface{} `json:"name,omitempty"`
	URI        string      `json:"uri,omitempty"`
}

// GlTFid ...
type GlTFid interface{}

// GltfMaterial ...
type GltfMaterial struct {
	AlphaCutoff          float64                      `json:"alphaCutoff,omitempty" validator:"gte=0"`
	AlphaMode            interface{}                  `json:"alphaMode,omitempty"`
	DoubleSided          bool                         `json:"doubleSided,omitempty"`
	EmissiveFactor       []float64                    `json:"emissiveFactor,omitempty"`
	EmissiveTexture      interface{}                  `json:"emissiveTexture,omitempty"`
	Extensions           interface{}                  `json:"extensions,omitempty"`
	Extras               interface{}                  `json:"extras,omitempty"`
	Name                 interface{}                  `json:"name,omitempty"`
	NormalTexture        interface{}                  `json:"normalTexture,omitempty"`
	OcclusionTexture     interface{}                  `json:"occlusionTexture,omitempty"`
	PbrMetallicRoughness MaterialPbrMetallicRoughness `json:"pbrMetallicRoughness,omitempty"`
}

// MaterialPbrMetallicRoughness ...
type MaterialPbrMetallicRoughness struct {
	BaseColorFactor          []float64   `json:"-"`
	BaseColorTexture         interface{} `json:"baseColorTexture,omitempty"`
	Extensions               interface{} `json:"extensions,omitempty"`
	Extras                   interface{} `json:"extras,omitempty"`
	MetallicFactor           float64     `json:"metallicFactor" validator:"gte=0, lte=1"`
	MetallicRoughnessTexture interface{} `json:"metallicRoughnessTexture,omitempty"`
	RoughnessFactor          float64     `json:"roughnessFactor,omitempty" validator:"gte=0, lte=1"`
}

// Mesh ...
type Mesh struct {
	Extensions interface{}     `json:"extensions,omitempty"`
	Extras     interface{}     `json:"extras,omitempty"`
	Name       string          `json:"name,omitempty"`
	Primitives []MeshPrimitive `json:"primitives,omitempty"`
	Weights    []float64       `json:"weights,omitempty"`
}

type meshInfoAssociation struct {
	MeshIndicesAccessorIndex     int
	MeshVerticesAccessorIndex    int
	MeshNormalsAccessorIndex     int
	MeshMaterialIndex            int
	MeshUVAccessorIndex          int
	MeshVertexColorAccessorIndex int
}

// MeshPrimitive ...
type MeshPrimitive struct {
	Attributes map[string]int `json:"attributes,omitempty"`
	Indices    int            `json:"indices" validator:"gte=0"`
	Material   int            `json:"material" validator:"gte=0"`
	Mode       int            `json:"mode,omitempty"`
}

// Node ...
type Node struct {
	Camera      interface{} `json:"camera,omitempty"`
	Children    []int       `json:"children,omitempty"`
	Extensions  interface{} `json:"extensions,omitempty"`
	Extras      interface{} `json:"extras,omitempty"`
	Matrix      []float64   `json:"matrix,omitempty"`
	Mesh        interface{} `json:"mesh,omitempty"`
	Name        string      `json:"name,omitempty"`
	Rotation    []float64   `json:"rotation,omitempty"`
	Scale       []float64   `json:"scale,omitempty"`
	Skin        interface{} `json:"skin,omitempty"`
	Translation []float64   `json:"translation,omitempty"`
	Weights     []float64   `json:"weights,omitempty"`
}

// Scene ...
type Scene struct {
	Extensions interface{} `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Name       interface{} `json:"name,omitempty"`
	Nodes      []int       `json:"nodes,omitempty"`
}
