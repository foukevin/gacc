package gacc

import (
	"os"
	"fmt"
	"encoding/json"
)

type Scene struct {
	Name string
}

type material struct {
	Name string
	Properties struct {
		Nofog bool `json:"no_fog"`
	}
	Colors []struct {
		Channel string
		Value [3]float64
	}
}

type meshData struct {
	Name string
}

type lightData struct {
	Color [3]float64
	Diffuse bool
	Distance float64
	Energy float64
	Specular bool
	Type string
}

type cameraData struct {
	Angle [2]float64
	Clipping [2]float64
	Type string
	Zoom []float64
}

type object struct {
	Name string
	Parent string
	Type string

	RawData json.RawMessage `json:"data"`
	//Data map[string]interface{}

	Transform struct {
		Location [3]float64
		Rotation [3]float64
		Scale [3]float64
		RotationOrder string `json:"rotation_order"`
	}
}

type jafScene struct {
	Contrib struct {
		AuthoringTool string `json:"authoring_tool"`
	}
	Info struct {
		Version float32
	} `json:"jaf"`
	Scene struct {
		Name string
		Materials []material
		Objects []object
	}
}

func ParseJafScene(filename string) Scene {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	dec := json.NewDecoder(file)

	var scene jafScene
	dec.Decode(&scene)
	fmt.Printf("%+v\n", scene)

	for _, o := range scene.Scene.Objects {
		switch (o.Type) {
		case "mesh":
			var data meshData
			json.Unmarshal(o.RawData, &data)
			fmt.Printf("%+v\n", data)
		case "light":
			var data lightData
			json.Unmarshal(o.RawData, &data)
			fmt.Printf("%+v\n", data)
		case "camera":
			var data cameraData
			json.Unmarshal(o.RawData, &data)
			fmt.Printf("%+v\n", data)
		}
	}

	return Scene{ Name: "untitled" }
}
