## Integrantes:
- Jhossep Martinez / 202173530-5
- Fernando Xais / 202273551-1
- Gabriela Yáñez / 202273511-2

## Consideraciones:
- Las imágenes ya están compiladas pero se puede ejecutar el siguiente comando en cada vm para volver a compilarlas: ```make build-<mv1|mv2|mv3|mv4>``` dependiendo de la vm en la que este.
- En el directorio de "/consumidores" se encuentra el archivo "consumidores.csv", desde aquí se leen las configuraciones de los consumidores, en caso de querer usar otro archivo debe mantenerse su ubicación y nombre.
- En el directorio de "/catalogos" se encuentran los archivos de catálogos de cada productor, en caso de querer usar otro archivo debe mantenerse su ubicación y nombre.
- Los consumidores almacenan sus ofertas en archivos csv dentro del directorio "/output" dentro de su respectiva máquina, en este directorio también se almacena el archivo Reporte.txt del broker.
- Si se cambia un archivo es necesario volver a ejecutar el build.
- Cuando se simula una caída, los nodos y consumidores responden con exito=false a las peticiones durante 5 segundos, luego solicitan resincronización al broker.
- Evitar utilizar Ctrl+C para terminar con la ejecución sin haber hecho uso del comando ```fin``` o ```reporte``` con anterioridad, puesto que el archivo reporte.txt no se generará y es posible que las entidades que se encontraban caídas sigan mostrando mensajes de recuperación automática. 

- Para poder ver los archivos generados dentro del directorio "/output" hay que ejecutar el siguiente comando en las máquinas respectivas:
 
   ~~~
   cat <nombreArchivo>
   ~~~


## Instrucciones:
- Abrir dos terminales de la VM dist16 y una terminal para dist13, dist14 y dist15. Ingresar al directorio "/lab2" en cada terminal

- Ir a la primera terminal de la VM dist16 y ejecutar ```make start-mv4```, esperar que se levanten los contenedores antes de continuar.
- Ir a la VM dist13 y ejecutar ```make start-mv1```.
- Ir a la VM dist14 y ejecutar ```make start-mv2```.
- Ir a la VM dist15 y ejecutar ```make start-mv3```.
- Cuando desee terminar con la ejecución ingrese a la segunda terminal de la VM dist16 y ejecute el siguiente comando ```make attach-broker```. Una vez se realice la conexión ejecute el comando ```fin``` o ```reporte``` para que los productores dejen de enviar ofertas y para que el broker pueda crear el reporte.
- Una vez generado el reporte se pueden terminar las ejecuciones con Ctrl+C, y se pueden ver los archivos generados ingresando al directorio "/output"

## Credenciales VM:
### dist013
- ehe6gqRsS2Fk
- 10.35.168.23
  
**Entidades:**
- Productor-riploy
- Nodo DB1
- Consumidores 5, 6, 7 y 8.

### dist014
- KRZ65kfAEmpB
- 10.35.168.24

**Entidades:**
- Productor-falabellox
- Nodo DB2
- Consumidores 9, 10, 11, 12.

### dist015
- aNASDGkYnQ8F
- 10.35.168.25

**Entidades:**
- Productor-parasio
- Nodo DB3

### dist016
- jrKU59Umn2TW
- 10.35.168.26

**Entidades:**
- Broker

- Consumidores 1, 2, 3, 4.
