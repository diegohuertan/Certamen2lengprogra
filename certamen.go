package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
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

func (d *dispatcher) ejecutarDispatcher(procesoActual chan bcp, tiempoEjecucion int) {

}

func leerArchivo(nombreArchivo string, lineas chan<- string, numeros chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	file, err := os.Open(nombreArchivo)
	if err != nil {
		fmt.Println("Error al abrir el archivo:", err)
		close(lineas)
		close(numeros)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		linea := scanner.Text()
		lineas <- linea

		words := strings.Fields(linea)
		if len(words) > 0 {
			numero, err := strconv.Atoi(words[0])
			if err == nil {
				for i := 1; i < len(words); i++ {
					numeros <- numero
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error al leer el archivo:", err)
	}

	close(lineas)
	close(numeros)
}

func main() {
	lineas := make(chan string)
	numeros := make(chan int)
	var wg sync.WaitGroup
	wg.Add(1)

	go leerArchivo("Orden_creacion.txt", lineas, numeros, &wg)
	var Tiempo_ejecucion []int
	var nombre_proceso []string

	var wgReaders sync.WaitGroup
	wgReaders.Add(2)

	go func() {
		defer wgReaders.Done()
		for numero := range numeros {
			Tiempo_ejecucion = append(Tiempo_ejecucion, numero)
		}
	}()
	go func() {
		defer wgReaders.Done()
		for linea := range lineas {
			words := strings.Fields(linea)
			// Guardar todos los procesos después del primer campo (tiempo)
			for i := 1; i < len(words); i++ {
				nombre_proceso = append(nombre_proceso, words[i])
			}
		}
	}()

	wg.Wait()
	wgReaders.Wait()
	indice := 0
	var d dispatcher
	fmt.Println("\n")

	for i := 1; i <= 5; i++ {
		fmt.Println("iteracion: ", d.contadorDisp)
		if indice < len(Tiempo_ejecucion) {
			for z := 0; z < len(Tiempo_ejecucion); z++ {
				x := Tiempo_ejecucion[z]
				if d.contadorDisp == x {
					d.agregarProceso(100+indice, "Listo", 0, nombre_proceso[indice])
					fmt.Println("Proceso", nombre_proceso[indice], "agregado a la cola de procesos")
					indice++
				}
			}
		}

		if i == 5 {
			if len(d.colaprocesos) > 1 {
				primero := d.colaprocesos[0]
				fmt.Println("primero", primero)
				d.colaprocesos = append(d.colaprocesos[1:], primero)
				fmt.Println("Proceso", d.colaprocesos[0])

			}
			i = 0
		}

		if d.contadorDisp == 15 {
			fmt.Println("cortando")
			break
		}
		d.contadorDisp++
	}
	// Pasar el primero al ultimo y el segundo al primero
	fmt.Println("\nNúmeros guardados en el slice:", Tiempo_ejecucion)
	fmt.Println("nombre guardados en el slice:", nombre_proceso)

	// Informacion del dispashe
	fmt.Println("Contenido de dispatcher:")
	for _, proceso := range d.colaprocesos {
		fmt.Printf("PID: %d, Estado: %s, ContadorProg: %d, NombreProceso: %s\n",
			proceso.pid, proceso.estado, proceso.contadorProg, proceso.nombreproceso)
		fmt.Println("Contador dispatcher: ", d.contadorDisp)
	}

}
