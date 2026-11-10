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
  String get homeWelcome => 'Tiquetera de almuerzos';

  @override
  String get errorNetwork =>
      'Sin conexión. Verifica tu red e inténtalo de nuevo.';

  @override
  String get errorTimeout => 'El servidor tardó demasiado en responder.';

  @override
  String get errorUnauthorized => 'Tu sesión expiró. Ingresa de nuevo.';

  @override
  String get errorUnknown => 'Ocurrió un error inesperado.';
}
