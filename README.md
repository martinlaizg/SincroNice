# SincroNice

# Requisitos
- Arqitectura cliente-servidor
    - Comunicación segura
        - Crear servidor HTTPS
    - Sistema de autenticación
        - Cifrado del login con SHA256
        - Soltier
        - Verificación en 2 pasos
        - Doble HASH (en el cliente y en el servidor)
        - Cifrar el HASH con una contraseña (guardar la contraseña en otra base de datos)
    - Almacenamiento
        - Esquema de almacenamiento
            - Simple
            - Incremental
            - o eliminación de bloques duplicados
        - Gestión de metadatos

## Requisitos mínimos
- Arquitectura cliente-servidor (comunicación sergura con terminales)
- Almacenamiento y recuperación de ficheros
- Sistema de autenticación seguro
- Cifrado de ficheros para su almacenamiento
- Logica de aplicación mínima

## Requisitos adicionales
- Doble autenticación
- Versión de ficheros
- Integración con otras APIs
- Sincronización automática
- Monitorización del sistema
- Cifrado con conocimiento cero (el servidor no conoce las contraseñas)

## Bibliografía
- [HTTPS Go Lang](https://www.kaihag.com/https-and-go/)
