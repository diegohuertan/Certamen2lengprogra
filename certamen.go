package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

type bcp struct {
	pid           int
	estado        string
	contadorProg  int
	nombreproceso string
}

type dispatcher struct {
	colaprocesos []bcp
	contadorDisp int
}

func (d *dispatcher) agregarProceso(pid int, estado string, contadorProg int, nombreproceso string) {
	d.colaprocesos = append(d.colaprocesos, bcp{pid, estado, contadorProg, nombreproceso})
}

func (d *dispatcher) ejecutarProceso(procesoActual chan bcp, tiempoEjecucion int) {
	for proceso := range procesoActual {
		nombreProceso := proceso.nombreproceso

		file, err := os.Open(nombreProceso)
		if err != nil {
			fmt.Println("Error al abrir el archivo:", err)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		linea := 0
		lineasEjecutadas := 0
		encontrado := false
		for scanner.Scan() {
			linea++
			if linea == proceso.contadorProg {
				encontrado = true
			}
			if encontrado {
				fmt.Println(scanner.Text())
				lineasEjecutadas++
				if lineasEjecutadas == tiempoEjecucion {
					proceso.contadorProg = linea
					fmt.Printf("Contador actualizado del proceso %d: %d\n", proceso.pid, proceso.contadorProg)
					d.contadorDisp = linea
					fmt.Printf("Contador del dispatcher actualizado: %d\n", d.contadorDisp)
					procesoActual <- proceso
					break
				}
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Error al leer el archivo:", err)
		}
	}
}

func leerArchivo(ruta string, lineas chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	file, err := os.Open(ruta)
	if err != nil {
		fmt.Println("Error al abrir el archivo:", err)
		close(lineas)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineas <- scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error al leer el archivo:", err)
	}

	close(lineas)
}

func main() {
	procesoActual := make(chan bcp)
	var wg sync.WaitGroup

	d := &dispatcher{}

	d.agregarProceso(1, "running", 1, "proceso_1")
	d.agregarProceso(2, "running", 6, "proceso_2")

	go func() {
		for _, proceso := range d.colaprocesos {
			procesoActual <- proceso
		}
		close(procesoActual)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		d.ejecutarProceso(procesoActual, 5)

	}()

	wg.Wait()
}
