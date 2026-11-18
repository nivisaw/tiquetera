// ignore: unused_import
import 'package:intl/intl.dart' as intl;

import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Spanish Castilian (`es`).
class AppLocalizationsEs extends AppLocalizations {
  AppLocalizationsEs([String locale = 'es']) : super(locale);

  @override
  String get appTitle => 'Tiquetera';

  @override
  String get loading => 'Cargando';

  @override
  String get fieldRequired => 'Campo obligatorio';

  @override
  String get loginUser => 'Usuario';

  @override
  String get loginPassword => 'Contraseña';

  @override
  String get loginSubmit => 'Ingresar';

  @override
  String get loginInvalidCredentials => 'Usuario o contraseña incorrectos.';

  @override
  String get loginRateLimited =>
      'Demasiados intentos. Espera un minuto e inténtalo de nuevo.';

  @override
  String get logout => 'Cerrar sesión';

  @override
  String get clientesTitle => 'Clientes';

  @override
  String get escanerTitle => 'Escáner';

  @override
  String get errorNetwork =>
      'Sin conexión. Verifica tu red e inténtalo de nuevo.';

  @override
  String get errorTimeout => 'El servidor tardó demasiado en responder.';

  @override
  String get errorUnauthorized => 'Tu sesión expiró. Ingresa de nuevo.';

  @override
  String get errorUnknown => 'Ocurrió un error inesperado.';

  @override
  String get retry => 'Reintentar';

  @override
  String get clientesSearch => 'Buscar por nombre o teléfono';

  @override
  String get clientesEmpty => 'No hay clientes registrados.';

  @override
  String get clientesNoResults =>
      'No hay clientes que coincidan con la búsqueda.';

  @override
  String get clientesNew => 'Nuevo cliente';

  @override
  String get clientesLoadMore => 'Cargar más';

  @override
  String get clienteInactivo => 'Inactivo';

  @override
  String get clienteNombre => 'Nombre';

  @override
  String get clienteTelefono => 'Teléfono';

  @override
  String get clienteDocumento => 'Documento (opcional)';

  @override
  String get clienteEmail => 'Correo (opcional)';

  @override
  String get clienteAlmuerzos => 'Almuerzos';

  @override
  String get clienteVencimiento => 'Vence';

  @override
  String get clienteSave => 'Guardar';

  @override
  String get clienteDecrease => 'Quitar un almuerzo';

  @override
  String get clienteIncrease => 'Agregar un almuerzo';

  @override
  String get clienteSelectDate => 'Elegir fecha de vencimiento';

  @override
  String get errNombre => 'Escribe entre 2 y 120 caracteres.';

  @override
  String get errTelefono => 'Escribe 10 dígitos, por ejemplo 300 123 4567.';

  @override
  String get errEmail => 'Correo no válido.';

  @override
  String get errDocumento => 'Máximo 30 caracteres.';

  @override
  String get errAlmuerzos => 'Entre 1 y 500 almuerzos.';

  @override
  String get errTelefonoDuplicado => 'Ya existe un cliente con ese teléfono.';

  @override
  String get errValidacion => 'Revisa los datos: algún valor no es válido.';

  @override
  String get errClienteInactivo => 'El cliente está inactivo.';

  @override
  String get errNotFound => 'El cliente no existe.';

  @override
  String get qrTitle => 'Código QR';

  @override
  String qrSemantics(String nombre) {
    return 'Código QR de $nombre';
  }

  @override
  String saldoLabel(int saldo) {
    String _temp0 = intl.Intl.pluralLogic(
      saldo,
      locale: localeName,
      other: '$saldo almuerzos disponibles',
      one: '1 almuerzo disponible',
    );
    return '$_temp0';
  }

  @override
  String venceLabel(String fecha) {
    return 'Vence el $fecha';
  }

  @override
  String get qrDone => 'Listo';

  @override
  String get qrShareWhatsApp => 'Compartir por WhatsApp';

  @override
  String get qrOpenChat => 'Abrir chat del cliente';

  @override
  String get qrShareError =>
      'No se pudo compartir el código QR. Intenta de nuevo.';

  @override
  String get qrChatError => 'No se pudo abrir WhatsApp.';

  @override
  String qrShareMessage(String nombre, int saldo, String fecha) {
    return 'Hola $nombre, esta es tu tiquetera de almuerzos. Saldo: $saldo almuerzos. Vence el $fecha. Presenta este código QR en caja.';
  }

  @override
  String qrChatMessage(String nombre) {
    return 'Hola $nombre, te envío el código QR de tu tiquetera de almuerzos.';
  }

  @override
  String get detalleTitle => 'Cliente';

  @override
  String get detalleSinTiquetera => 'Sin tiquetera activa.';

  @override
  String get detalleRecargar => 'Recargar';

  @override
  String get detalleRegenerar => 'Regenerar QR';

  @override
  String get historialTitle => 'Historial de consumos';

  @override
  String get historialEmpty => 'Aún no hay consumos registrados.';

  @override
  String get historialError => 'No se pudo cargar el historial.';

  @override
  String historialSaldo(int saldo) {
    String _temp0 = intl.Intl.pluralLogic(
      saldo,
      locale: localeName,
      other: 'Quedaron $saldo almuerzos',
      one: 'Quedó 1 almuerzo',
    );
    return '$_temp0';
  }

  @override
  String get recargaTitle => 'Nueva recarga';

  @override
  String get recargaHint => 'Se suma al saldo actual.';

  @override
  String get recargaOk => 'Recarga registrada.';

  @override
  String get cancel => 'Cancelar';

  @override
  String get regenerarTitle => '¿Regenerar el código QR?';

  @override
  String get regenerarBody =>
      'El código QR anterior dejará de funcionar. El cliente necesitará el nuevo.';

  @override
  String get regenerarConfirm => 'Regenerar';

  @override
  String get escanerPermisoTitle => 'Necesitamos la cámara';

  @override
  String get escanerPermisoBody =>
      'La cámara se usa solo para leer el código QR de la tiquetera del cliente.';

  @override
  String get escanerPermisoAllow => 'Permitir cámara';

  @override
  String get escanerPermisoDenied =>
      'El permiso de cámara está bloqueado. Actívalo en los ajustes del teléfono.';

  @override
  String get escanerPermisoSettings => 'Abrir ajustes';

  @override
  String get escanerGuia => 'Apunta al código QR del cliente';

  @override
  String get escanerLinterna => 'Linterna';

  @override
  String get escanerProcesando => 'Registrando almuerzo…';

  @override
  String get escanerExito => '¡Almuerzo registrado!';

  @override
  String get escanerSaldoLabel => 'almuerzos restantes';

  @override
  String escanerVence(String fecha) {
    return 'Vence el $fecha';
  }

  @override
  String get escanerTocaParaVolver => 'Toca para escanear otro';

  @override
  String get escanerRechazoTitle => 'Almuerzo no registrado';

  @override
  String get rechazoQrInvalido => 'Este código no es una tiquetera válida.';

  @override
  String get rechazoClienteNoExiste =>
      'Código desconocido o reemplazado. Pide al cliente su QR más reciente.';

  @override
  String get rechazoClienteInactivo => 'El cliente está inactivo.';

  @override
  String get rechazoSinTiquetera => 'El cliente no tiene una tiquetera activa.';

  @override
  String get rechazoTiqueteraVencida => 'La tiquetera está vencida.';

  @override
  String get rechazoSinSaldo => 'La tiquetera no tiene almuerzos disponibles.';

  @override
  String get rechazoSinRed => 'Sin conexión con el servidor. Escanea de nuevo.';

  @override
  String get rechazoDemasiados => 'Demasiados escaneos. Espera un minuto.';

  @override
  String get rechazoError => 'No se pudo registrar. Inténtalo de nuevo.';
}
