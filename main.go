package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/pipeline/pkg/client/clientset/versioned/scheme"
)

var (
	filename = flag.String("f", "", "file to parse or directory that will be walked to find tasks definition")
)

// TaskElement is a structure that contains bot the fileinfo and the task
type TaskElement struct {
	Fileinfo os.FileInfo
	Task     v1beta1.Task
}

func main() {
	flag.Parse()

	fileinfo, err := os.Stat(*filename)
	if err != nil {
		log.Fatal(err)
		return
	}
	var taskElements []TaskElement
	if fileinfo.IsDir() {
		files, err := ioutil.ReadDir(*filename)
		if err != nil {
			log.Fatal(err)
			return
		}
		for _, file := range files {
			if strings.HasPrefix(file.Name(), "task-") {
				taskElements = append(taskElements, visitTaskFile(path.Join(*filename, file.Name()), file))
			}
		}
	} else {
		taskElements = append(taskElements, visitTaskFile(*filename, fileinfo))
	}

	generateDoc(filepath.Dir(*filename), taskElements)
}

func visitTaskFile(path string, info os.FileInfo) TaskElement {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("error reading file: %v", err)
	}

	var task v1beta1.Task

	if _, _, err := scheme.Codecs.UniversalDeserializer().Decode(dat, nil, &task); err != nil {
		log.Fatal(err)
	}

	return TaskElement{info, task}

}

func generateDoc(folder string, taskElements []TaskElement) {

	var buff bytes.Buffer

	buff.WriteString("# " + folder + " related tasks")
	summary := `{{range .}}-**[{{.Task.Name}}](#{{.Task.Name}})**: {{.Task.Spec.Description}}
{{end}}`
	tmpl := template.Must(template.New("summary").Parse(summary))
	if err := tmpl.Execute(&buff, taskElements); err != nil {
		log.Fatalf("error executing the template: %v", err)
	}

	buff.WriteString("## Install the Tasks")
	buff.WriteString("- Add a github integration in your toolchain to the repository containing the task (https://github.com/open-toolchain/tekton-catalog)")
	buff.WriteString("- Add that github integration to the Definitions tab of your Continuous Delivery tekton pipeline, with the Path set to `" + folder + "`")

	for _, taskElement := range taskElements {

		var task = taskElement.Task

		tmpl := template.Must(template.New("test").Parse(`# {{.Name}}
## Install the Task
kubectl apply -f https://raw.githubusercontent.com/tektoncd/catalog/master/{{.Name}}/{{.Name}}.yaml
### Input:-
`))
		if err := tmpl.Execute(&buff, task); err != nil {
			log.Fatalf("error executing the template: %v", err)
		}

		if task.Spec.Params != nil {
			t := `{{range .}}
- {{.Name}}, {{.Description}}
{{end}}`
			tmpl := template.Must(template.New("test").Parse(t))
			if err := tmpl.Execute(&buff, task.Spec.Params); err != nil {
				log.Fatalf("error executing the template: %v", err)
			}
		}

		if task.Spec.Resources != nil {

			if task.Spec.Resources.Inputs != nil {
				t := `### Input:- {{range .}}- {{.ResourceDeclaration.Name}}, {{.ResourceDeclaration.Type}}	{{end}}`
				tmpl := template.Must(template.New("test").Parse(t))
				if err := tmpl.Execute(&buff, task.Spec.Resources.Inputs); err != nil {
					log.Fatalf("error executing the template: %v", err)
				}
			}

			if task.Spec.Resources.Outputs != nil {
				t := `### Output:- {{range .}}- {{.ResourceDeclaration.Name}}, {{.ResourceDeclaration.Type}} {{end}}`
				tmpl := template.Must(template.New("test").Parse(t))
				err := tmpl.Execute(&buff, task.Spec.Resources.Outputs)
				if err != nil {
					log.Fatalf("error executing the template: %v", err)
				}
			}
		}
	}

	fmt.Print(buff.String())
}
