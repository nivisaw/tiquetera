# Tiquetera de Almuerzos

App móvil (Android/iOS) para gestionar tiqueteras prepagadas de almuerzos con QR.

- **Admin:** registra clientes, asigna y recarga paquetes, genera y comparte el QR.
- **Cajero:** escanea el QR y descuenta un almuerzo.

| Carpeta | Contenido |
|---|---|
| `backend/` | API Go (chi, pgx, sqlc) |
| `mobile/` | App Flutter |
| `infra/` | Docker Compose y variables de entorno (`.env.example`) |
| `docs/` | Arquitectura, modelo de datos, contrato API, ADR, plan y avance |

Estado del desarrollo: [docs/PROGRESS.md](docs/PROGRESS.md). Contrato de la API: [docs/API.md](docs/API.md).

## Requisitos

- Docker Desktop (con Docker Compose v2). Todo el backend corre en contenedores.
- Flutter SDK 3.47 y Android Studio con un emulador (ej. `Medium_Phone_API_37.0`), o un teléfono Android con depuración USB.
- Git Bash en Windows. `make` es opcional: abajo están los comandos equivalentes de Docker.

## Accesos

### Usuarios de prueba (solo entorno local)

| Rol | Usuario | Contraseña |
|---|---|---|
| Admin | `prueba.admin` | `Prueba2026!` |
| Cajero | `prueba.cajero` | `Cajero2026!` |

Existen solo en la base de datos local. Si se borran los datos (`$C down -v`), se recrean con el script de [Crear usuarios de prueba](#crear-usuarios-de-prueba). Nunca usar estas cuentas ni contraseñas en producción.

Tras 5 intentos fallidos la cuenta queda bloqueada 15 minutos (el mensaje es el mismo que el de contraseña incorrecta).

### Configuración del entorno

Las credenciales **no se guardan en el repositorio**. Se definen en `infra/.env` (ignorado por git), que se crea a partir de [infra/.env.example](infra/.env.example).

| Acceso | Dónde se define | Valor local por defecto |
|---|---|---|
| Admin de la app | `SEED_ADMIN_USERNAME` / `SEED_ADMIN_PASSWORD` | usuario `admin`; contraseña: la que pongas (≥ 10 caracteres) |
| Cajero de la app | Se crea con la API (ver [Crear usuarios de prueba](#crear-usuarios-de-prueba)) | `prueba.cajero` |
| API local | `API_HOST_PORT` | `http://127.0.0.1:8088/v1` (desde el emulador: `http://10.0.2.2:8088/v1`) |
| PostgreSQL local (owner) | `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_HOST_PORT` | `tiquetera_owner` en `127.0.0.1:5442`, base `tiquetera` |
| PostgreSQL (rol de la API) | `APP_DB_PASSWORD` | rol `tiquetera_app` (sin privilegios de borrado) |
| Firma JWT | `JWT_PRIVATE_KEY` | generar con `openssl rand -base64 32` |

Reemplaza **todos** los valores `change-me-*` antes de levantar el stack. En producción las contraseñas deben ser distintas y fuertes.

## Levantar el backend

Desde la raíz del repo, en Git Bash:

```bash
cp -n infra/.env.example infra/.env        # luego edita los valores change-me-*
C="docker compose --env-file infra/.env -f infra/docker-compose.yml"

$C up -d --wait                             # db + api           (make up)
$C run --rm migrate up                      # migraciones        (make migrate)
$C run --rm db-app-user                     # contraseña del rol de la API
$C run --rm --build seed                    # admin inicial      (make seed)
$C exec -T db sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB"' < infra/scripts/seed_demo.sql   # 40 clientes demo (make seed-demo)

curl -s -o /dev/null -w "%{http_code}\n" 127.0.0.1:8088/readyz   # 200
```

Otros comandos:

| Acción | Comando | make |
|---|---|---|
| Ver estado | `$C ps` | `make ps` |
| Ver logs | `$C logs -f api` | `make logs` |
| Detener | `$C down` | `make down` |
| Borrar datos locales | `$C down -v` | `make clean` |

### Crear usuarios de prueba

Usa el admin inicial de `infra/.env` para crear las cuentas de prueba (responde `201`, o `409` si ya existen):

```bash
P=$(grep ^SEED_ADMIN_PASSWORD= infra/.env | cut -d= -f2-)
TOKEN=$(curl -s 127.0.0.1:8088/v1/auth/login -H 'Content-Type: application/json' \
  -d "{\"username\":\"admin\",\"password\":\"$P\"}" | sed -E 's/.*"access_token":"([^"]+)".*/\1/')

curl -s 127.0.0.1:8088/v1/usuarios -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"username":"prueba.admin","password":"Prueba2026!","nombre":"Admin Prueba","rol":"admin"}'
curl -s 127.0.0.1:8088/v1/usuarios -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"username":"prueba.cajero","password":"Cajero2026!","nombre":"Cajero Prueba","rol":"cajero"}'
```

## Ejecutar la app móvil

```bash
cd mobile
flutter pub get
flutter emulators --launch Medium_Phone_API_37.0
flutter run --dart-define=API_BASE_URL=http://10.0.2.2:8088/v1 --dart-define=FLAVOR=dev
```

- `API_BASE_URL` debe terminar en `/v1`. En un teléfono físico usa la IP de tu PC en la red local (ej. `http://192.168.1.20:8088/v1`). Para eso la API debe escuchar en esa interfaz, no solo en `127.0.0.1`.
- `FLAVOR=prod` exige `https`.
- Si el emulador muestra la pantalla negra (problema de GPU con Impeller), agrega `--no-enable-impeller`.
- La primera compilación tarda 1-3 minutos. Espera hasta ver `Flutter run key commands`. Con la app abierta, `r` recarga los cambios y `q` la cierra.
- Si `flutter emulators --launch` muestra "exited with code 1", normalmente el emulador ya estaba abierto. Confírmalo con `flutter devices`: debe aparecer `emulator-5554`.
- Antes de abrir la app, verifica que el backend responda: `curl 127.0.0.1:8088/readyz` debe devolver `{"status":"ready"}`.

### Avisos de consola que se pueden ignorar

- `WARNING: Restricted methods will be blocked in a future release unless native access is enabled` (y las líneas `WARNING: java.lang.System::load ...` que lo acompañan) los imprime Gradle al usar un JDK 24 o superior. Son solo informativos: no impiden compilar ni ejecutar la app.
- Las líneas `hiddenapi: Accessing hidden method ...` y `Width is zero` de logcat también son normales.

## Pruebas automáticas

```bash
# Backend (lint + tests con testcontainers; requiere Docker)
make lint test

# Móvil
cd mobile
flutter analyze
flutter test
flutter test --coverage                     # informe en mobile/coverage/lcov.info
flutter test --update-goldens               # solo si cambió a propósito una pantalla con golden
```

Los goldens se generan en Windows y CI corre en Linux; se tolera hasta 1 % de diferencia de píxeles ([mobile/test/flutter_test_config.dart](mobile/test/flutter_test_config.dart)).

CI (GitHub Actions) ejecuta en cada PR: backend lint, test, govulncheck y gosec; mobile analyze y test; build de imagen y trivy.

## Prueba manual (admin)

Con el backend arriba y la app en el emulador, entra como `prueba.admin` / `Prueba2026!`. El paso 1 requiere una base sin clientes; si ya hay, ejecuta `$C down -v` y vuelve a [levantar el backend](#levantar-el-backend) y a [crear los usuarios de prueba](#crear-usuarios-de-prueba).

| # | Acción | Esperado |
|---|---|---|
| 1 | Abrir Clientes sin datos | "No hay clientes registrados." |
| 2 | "Nuevo cliente" → Guardar vacío | Errores en Nombre y Teléfono; no envía |
| 3 | Teléfono `123`, correo `ana@`, almuerzos `501` | Error en línea en cada campo |
| 4 | Botones − / + de almuerzos | − deshabilitado en 1; + deshabilitado en 500 |
| 5 | Selector de fecha | No permite días anteriores a hoy; formato dd/MM/yyyy |
| 6 | Datos válidos (teléfono `300 123 4567`) → Guardar | Spinner y botón deshabilitado; luego pantalla QR con nombre, saldo y vencimiento |
| 7 | Pantalla QR | El token no aparece en texto; QR negro sobre blanco también en modo oscuro |
| 8 | "Listo" | Abre el detalle del cliente |
| 9 | Crear otro cliente con el mismo teléfono | "Ya existe un cliente con ese teléfono." |
| 10 | Buscar por nombre o teléfono en el listado | Filtra; sin coincidencias muestra el mensaje de "sin resultados" |
| 11 | Detalle → Recargar 20 | "Recarga registrada."; el saldo suma la recarga |
| 12 | Regenerar QR → Cancelar | No cambia nada |
| 13 | Regenerar QR → Confirmar | Muestra un QR nuevo; el anterior deja de servir |
| 14 | `$C stop api` y recargar el listado | Mensaje de sin conexión y botón "Reintentar"; funciona otra vez tras `$C start api` |
| 15 | Texto al 200 % y modo oscuro en Ajustes de Android | Sin desbordes, legible |

Al terminar, verifica que la API no registre tokens ni payloads del QR:

```bash
$C logs api | grep -c "TQ1\."              # esperado: 0
```

El flujo del cajero (escáner y débito) se agrega en la sesión S12.
