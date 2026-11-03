# Reglas UX/UI

## Principios
- Material 3, un solo `ThemeData` con tokens (colores, espaciado 4/8/16/24, radios). Soporte modo oscuro.
- Uso en caja con prisa: acciones principales con objetivos táctiles ≥ 48dp, máximo 2 toques para la tarea central.
- Toda pantalla con estados: cargando (skeleton), vacío, error (con reintentar), éxito.
- Accesibilidad: contraste WCAG AA, `Semantics` en iconos, soporte de texto escalado hasta 200% sin desbordes. El color nunca es la única señal (icono + texto).

## Cajero
- Escáner a pantalla completa con marco guía, botón linterna.
- **Éxito:** fondo verde (#1B873F aprox.), check animado grande, "¡Almuerzo registrado!", nombre del cliente, **saldo restante ≥ 64sp**, vencimiento. Vibración corta + sonido.
- **Rechazo:** fondo rojo (#C62828 aprox.), icono X, motivo en lenguaje claro (ej. "Tiquetera vencida el 30/10/2026"), vibración doble.
- Auto-retorno al escáner en 3 s o tocando la pantalla.

## Admin
- Formulario de una columna, teclado adecuado por campo (`phone`, `number`), validación en línea, botón Guardar deshabilitado durante el envío.
- Pantalla QR: QR grande con alto contraste y margen blanco, nombre y saldo debajo, botones "Compartir por WhatsApp" (primario) y "Otras apps".
- Confirmación antes de acciones destructivas (regenerar QR invalida el anterior).

## Formatos
- Fechas `dd/MM/yyyy`, locale `es_CO`. Teléfonos mostrados con formato local.
