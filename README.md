# Knis

Este proyecto es una aplicación escrita en Go que obtiene información de personas a partir de su DNI, y almacena esta información en una base de datos MongoDB. La aplicación se ejecuta dentro de un contenedor Docker y utiliza `docker-compose` para manejar la configuración y el despliegue de los servicios.

## Requisitos Previos

Antes de empezar, asegúrate de tener instalados los siguientes programas en tu máquina:

- **Docker**: [Instalación de Docker](https://docs.docker.com/get-docker/)
- **Docker Compose**: [Instalación de Docker Compose](https://docs.docker.com/compose/install/)

## Configuración del Proyecto

1. Clona este repositorio en tu máquina local:

    ```bash
    git clone https://github.com/keynertyc/knis.git
    cd knis
    ```

2. Asegúrate de que tienes un archivo `docker-compose.yml` en la raíz del proyecto. Este archivo contiene la configuración para lanzar los servicios necesarios, como MongoDB y la aplicación Go.

3. Verifica que el archivo `Dockerfile` está configurado correctamente para construir la imagen de la aplicación.

## Configuración del Entorno

El archivo `docker-compose.yml` define dos servicios:

- **MongoDB**: Se ejecuta utilizando la imagen oficial de MongoDB y expone el puerto `27017`.
- **Aplicación Go**: La aplicación se construye y ejecuta utilizando el contenedor Go, y depende del servicio MongoDB para almacenar los datos.

### Variables de Entorno

Las siguientes variables de entorno se configuran en el servicio de la aplicación Go:

- `MONGO_URI`: La URI de conexión a MongoDB (por defecto `mongodb://mongo:27017`).
- `DB_NAME`: El nombre de la base de datos en MongoDB (por defecto `knisdb`).

## Construcción y Ejecución

Para construir y ejecutar el proyecto, sigue estos pasos:

1. Construye y levanta los servicios utilizando `docker-compose`:

    ```bash
    docker compose up --build -d
    ```

   Este comando descargará las imágenes necesarias, construirá la imagen de la aplicación y levantará tanto MongoDB como la aplicación Go.

2. La aplicación estará en modo "infinito" debido al comando `CMD ["sleep", "infinity"]` en el `Dockerfile`. Para ejecutar la aplicación manualmente, puedes acceder al contenedor de la aplicación:

    ```bash
    docker exec -it go_app /bin/bash
    ```

   Y dentro del contenedor, puedes ejecutar el binario de la aplicación:

    ```bash
    ./main 40000000 49999999
    ```

   Esto iniciará la aplicación para procesar el rango de DNIs especificado.

## Volúmenes

Se utilizan volúmenes Docker para persistir datos:

- **mongo-data**: Almacena los datos de MongoDB de forma persistente.
- **app_data**: Almacena los datos generados por la aplicación dentro del directorio `/app/data`.

## Limpieza

Para detener los contenedores y eliminar los volúmenes persistentes, puedes ejecutar:

```bash
docker-compose down -v
```

Este comando detendrá y eliminará todos los contenedores y volúmenes asociados al proyecto.

## Contribuciones

Las contribuciones son bienvenidas. Puedes abrir un issue o un pull request en este repositorio.

## Licencia

Este proyecto está licenciado bajo la Licencia MIT. Ver [LICENSE](LICENSE) para más detalles.

## Disclaimer

Este proyecto se proporciona con fines educativos y de demostración. El desarrollador **no se responsabiliza** del mal uso de esta herramienta, incluyendo, pero no limitado a, la violación de leyes de privacidad, uso indebido de la información obtenida, o cualquier otra actividad que pueda considerarse ilegal o inapropiada. Es responsabilidad del usuario asegurarse de cumplir con todas las leyes y regulaciones aplicables al utilizar este software.

## Autor

- **Keyner TYC**  
  Email: [keyner.peru@gmail.com](mailto:keyner.peru@gmail.com)