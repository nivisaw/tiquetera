# Reglas Móvil Flutter

- Dart 3, null-safety, `very_good_analysis`; `flutter analyze` con 0 issues.
- Arquitectura feature-first (`data/ domain/ presentation/`). Widgets sin lógica de red; la lógica vive en Notifiers/repositorios.
- Estado: Riverpod (`riverpod_generator`). Inmutabilidad con `freezed`; JSON con `json_serializable`.
- Red: un único cliente dio; interceptores de auth/refresh (con lock para evitar refresh concurrente) y mapeo de errores a `AppFailure` sellada. Timeouts 10 s.
- `const` constructors siempre que sea posible; listas con `ListView.builder`; evitar rebuilds (select en providers).
- Configuración por `--dart-define` (API_BASE_URL, FLAVOR). Nada hardcodeado.
- Paquetes aprobados: flutter_riverpod, go_router, dio, freezed, json_serializable, flutter_secure_storage, mobile_scanner, qr_flutter, share_plus, path_provider, url_launcher, intl, permission_handler, mocktail. Otro → justificar.
- Permisos: solicitar cámara solo al entrar al escáner, con pantalla explicativa.
- Textos de UI en español vía `intl` (ARB), nunca literales en widgets.
