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
- Versión de ficheros
- Integración con otras APIs

## Requisitos adicionales
- Doble autenticación
- Sincronización automática
- Monitorización del sistema
- Cifrado con conocimiento cero (el servidor no conoce las contraseñas)

## Bibliografía
- [HTTPS Go Lang](https://www.kaihag.com/https-and-go/)




# Pasos de seguridad
1. Cliente realiza hash (sha256) sobre su contraseña.
2. Se transmite el paquete mediante HTTPS.
3. El servidor realiza el hash (bcrypt o scrypt) con salt.
4. El servidor cifra la contraseña (AES) para almacenarla en el servidor.

## Opcional
- Con el token de verificación en dos pasos, usarlo para mantener la sesión.
