package main

import (
	"bufio"
	"flag"
	"log"
	"os/exec"
	"strings"
	"sync"
)

func main() {
	// Define and parse flags
	kFlag := flag.Bool("k", false, "Kill non-running pods")
	nFlag := flag.String("n", "", "Pod namespace")
	flag.Parse()

	// Flag actions
	if *kFlag {
		killNonRunningPods(*nFlag)
	}
}

func killNonRunningPods(namespace string) {
	pods := getPodList(namespace)

	var wg sync.WaitGroup
	for _, pod := range pods {
		cols := strings.Fields(pod)
		if len(cols) == 5 && cols[2] != "Running" {
			wg.Add(1)
			go killPod(cols[0], namespace, &wg)
		}
	}
	wg.Wait()
}

func killPod(podID, namespace string, wg *sync.WaitGroup) {
	defer wg.Done()

	var cmdArgs = []string{"delete", "pod", podID}
	if namespace != "" {
		cmdArgs = append(cmdArgs, "-n", namespace)
	}
	log.Println("Deleting pod:", podID)
	cmd := exec.Command("kubectl", cmdArgs...)
	cmd.Run()
}

func getPodList(namespace string) []string {

	var cmdArgs = []string{"get", "pods"}
	if namespace != "" {
		cmdArgs = append(cmdArgs, "-n", namespace)
	}

	cmd := exec.Command("kubectl", cmdArgs...)

	r, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	scanner := bufio.NewScanner(r)

	var pods []string
	done := make(chan bool)

	go func() {

		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "NAME") {
				continue
			}
			pods = append(pods, line)
		}

		done <- true

	}()

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	<-done

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}

	return pods
}
