import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_es.dart';

// ignore_for_file: type=lint

/// Callers can lookup localized strings with an instance of AppLocalizations
/// returned by `AppLocalizations.of(context)`.
///
/// Applications need to include `AppLocalizations.delegate()` in their app's
/// `localizationDelegates` list, and the locales they support in the app's
/// `supportedLocales` list. For example:
///
/// ```dart
/// import 'gen/app_localizations.dart';
///
/// return MaterialApp(
///   localizationsDelegates: AppLocalizations.localizationsDelegates,
///   supportedLocales: AppLocalizations.supportedLocales,
///   home: MyApplicationHome(),
/// );
/// ```
///
/// ## Update pubspec.yaml
///
/// Please make sure to update your pubspec.yaml to include the following
/// packages:
///
/// ```yaml
/// dependencies:
///   # Internationalization support.
///   flutter_localizations:
///     sdk: flutter
///   intl: any # Use the pinned version from flutter_localizations
///
///   # Rest of dependencies
/// ```
///
/// ## iOS Applications
///
/// iOS applications define key application metadata, including supported
/// locales, in an Info.plist file that is built into the application bundle.
/// To configure the locales supported by your app, you’ll need to edit this
/// file.
///
/// First, open your project’s ios/Runner.xcworkspace Xcode workspace file.
/// Then, in the Project Navigator, open the Info.plist file under the Runner
/// project’s Runner folder.
///
/// Next, select the Information Property List item, select Add Item from the
/// Editor menu, then select Localizations from the pop-up menu.
///
/// Select and expand the newly-created Localizations item then, for each
/// locale your application supports, add a new item and select the locale
/// you wish to add from the pop-up menu in the Value field. This list should
/// be consistent with the languages listed in the AppLocalizations.supportedLocales
/// property.
abstract class AppLocalizations {
  AppLocalizations(String locale)
    : localeName = intl.Intl.canonicalizedLocale(locale.toString());

  final String localeName;

  static AppLocalizations of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations)!;
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  /// A list of this localizations delegate along with the default localizations
  /// delegates.
  ///
  /// Returns a list of localizations delegates containing this delegate along with
  /// GlobalMaterialLocalizations.delegate, GlobalCupertinoLocalizations.delegate,
  /// and GlobalWidgetsLocalizations.delegate.
  ///
  /// Additional delegates can be added by appending to this list in
  /// MaterialApp. This list does not have to be used at all if a custom list
  /// of delegates is preferred or required.
  static const List<LocalizationsDelegate<dynamic>> localizationsDelegates =
      <LocalizationsDelegate<dynamic>>[
        delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
      ];

  /// A list of this localizations delegate's supported locales.
  static const List<Locale> supportedLocales = <Locale>[Locale('es')];

  /// No description provided for @appTitle.
  ///
  /// In es, this message translates to:
  /// **'Tiquetera'**
  String get appTitle;

  /// No description provided for @loading.
  ///
  /// In es, this message translates to:
  /// **'Cargando'**
  String get loading;

  /// No description provided for @fieldRequired.
  ///
  /// In es, this message translates to:
  /// **'Campo obligatorio'**
  String get fieldRequired;

  /// No description provided for @loginUser.
  ///
  /// In es, this message translates to:
  /// **'Usuario'**
  String get loginUser;

  /// No description provided for @loginPassword.
  ///
  /// In es, this message translates to:
  /// **'Contraseña'**
  String get loginPassword;

  /// No description provided for @loginSubmit.
  ///
  /// In es, this message translates to:
  /// **'Ingresar'**
  String get loginSubmit;

  /// No description provided for @loginInvalidCredentials.
  ///
  /// In es, this message translates to:
  /// **'Usuario o contraseña incorrectos.'**
  String get loginInvalidCredentials;

  /// No description provided for @loginRateLimited.
  ///
  /// In es, this message translates to:
  /// **'Demasiados intentos. Espera un minuto e inténtalo de nuevo.'**
  String get loginRateLimited;

  /// No description provided for @logout.
  ///
  /// In es, this message translates to:
  /// **'Cerrar sesión'**
  String get logout;

  /// No description provided for @clientesTitle.
  ///
  /// In es, this message translates to:
  /// **'Clientes'**
  String get clientesTitle;

  /// No description provided for @escanerTitle.
  ///
  /// In es, this message translates to:
  /// **'Escáner'**
  String get escanerTitle;

  /// No description provided for @errorNetwork.
  ///
  /// In es, this message translates to:
  /// **'Sin conexión. Verifica tu red e inténtalo de nuevo.'**
  String get errorNetwork;

  /// No description provided for @errorTimeout.
  ///
  /// In es, this message translates to:
  /// **'El servidor tardó demasiado en responder.'**
  String get errorTimeout;

  /// No description provided for @errorUnauthorized.
  ///
  /// In es, this message translates to:
  /// **'Tu sesión expiró. Ingresa de nuevo.'**
  String get errorUnauthorized;

  /// No description provided for @errorUnknown.
  ///
  /// In es, this message translates to:
  /// **'Ocurrió un error inesperado.'**
  String get errorUnknown;

  /// No description provided for @retry.
  ///
  /// In es, this message translates to:
  /// **'Reintentar'**
  String get retry;

  /// No description provided for @clientesSearch.
  ///
  /// In es, this message translates to:
  /// **'Buscar por nombre o teléfono'**
  String get clientesSearch;

  /// No description provided for @clientesEmpty.
  ///
  /// In es, this message translates to:
  /// **'No hay clientes registrados.'**
  String get clientesEmpty;

  /// No description provided for @clientesNoResults.
  ///
  /// In es, this message translates to:
  /// **'No hay clientes que coincidan con la búsqueda.'**
  String get clientesNoResults;

  /// No description provided for @clientesNew.
  ///
  /// In es, this message translates to:
  /// **'Nuevo cliente'**
  String get clientesNew;

  /// No description provided for @clientesLoadMore.
  ///
  /// In es, this message translates to:
  /// **'Cargar más'**
  String get clientesLoadMore;

  /// No description provided for @clienteInactivo.
  ///
  /// In es, this message translates to:
  /// **'Inactivo'**
  String get clienteInactivo;

  /// No description provided for @clienteNombre.
  ///
  /// In es, this message translates to:
  /// **'Nombre'**
  String get clienteNombre;

  /// No description provided for @clienteTelefono.
  ///
  /// In es, this message translates to:
  /// **'Teléfono'**
  String get clienteTelefono;

  /// No description provided for @clienteDocumento.
  ///
  /// In es, this message translates to:
  /// **'Documento (opcional)'**
  String get clienteDocumento;

  /// No description provided for @clienteEmail.
  ///
  /// In es, this message translates to:
  /// **'Correo (opcional)'**
  String get clienteEmail;

  /// No description provided for @clienteAlmuerzos.
  ///
  /// In es, this message translates to:
  /// **'Almuerzos'**
  String get clienteAlmuerzos;

  /// No description provided for @clienteVencimiento.
  ///
  /// In es, this message translates to:
  /// **'Vence'**
  String get clienteVencimiento;

  /// No description provided for @clienteSave.
  ///
  /// In es, this message translates to:
  /// **'Guardar'**
  String get clienteSave;

  /// No description provided for @clienteDecrease.
  ///
  /// In es, this message translates to:
  /// **'Quitar un almuerzo'**
  String get clienteDecrease;

  /// No description provided for @clienteIncrease.
  ///
  /// In es, this message translates to:
  /// **'Agregar un almuerzo'**
  String get clienteIncrease;

  /// No description provided for @clienteSelectDate.
  ///
  /// In es, this message translates to:
  /// **'Elegir fecha de vencimiento'**
  String get clienteSelectDate;

  /// No description provided for @errNombre.
  ///
  /// In es, this message translates to:
  /// **'Escribe entre 2 y 120 caracteres.'**
  String get errNombre;

  /// No description provided for @errTelefono.
  ///
  /// In es, this message translates to:
  /// **'Escribe 10 dígitos, por ejemplo 300 123 4567.'**
  String get errTelefono;

  /// No description provided for @errEmail.
  ///
  /// In es, this message translates to:
  /// **'Correo no válido.'**
  String get errEmail;

  /// No description provided for @errDocumento.
  ///
  /// In es, this message translates to:
  /// **'Máximo 30 caracteres.'**
  String get errDocumento;

  /// No description provided for @errAlmuerzos.
  ///
  /// In es, this message translates to:
  /// **'Entre 1 y 500 almuerzos.'**
  String get errAlmuerzos;

  /// No description provided for @errTelefonoDuplicado.
  ///
  /// In es, this message translates to:
  /// **'Ya existe un cliente con ese teléfono.'**
  String get errTelefonoDuplicado;

  /// No description provided for @errValidacion.
  ///
  /// In es, this message translates to:
  /// **'Revisa los datos: algún valor no es válido.'**
  String get errValidacion;

  /// No description provided for @errClienteInactivo.
  ///
  /// In es, this message translates to:
  /// **'El cliente está inactivo.'**
  String get errClienteInactivo;

  /// No description provided for @errNotFound.
  ///
  /// In es, this message translates to:
  /// **'El cliente no existe.'**
  String get errNotFound;

  /// No description provided for @qrTitle.
  ///
  /// In es, this message translates to:
  /// **'Código QR'**
  String get qrTitle;

  /// No description provided for @qrSemantics.
  ///
  /// In es, this message translates to:
  /// **'Código QR de {nombre}'**
  String qrSemantics(String nombre);

  /// No description provided for @saldoLabel.
  ///
  /// In es, this message translates to:
  /// **'{saldo, plural, =1{1 almuerzo disponible} other{{saldo} almuerzos disponibles}}'**
  String saldoLabel(int saldo);

  /// No description provided for @venceLabel.
  ///
  /// In es, this message translates to:
  /// **'Vence el {fecha}'**
  String venceLabel(String fecha);

  /// No description provided for @qrDone.
  ///
  /// In es, this message translates to:
  /// **'Listo'**
  String get qrDone;

  /// No description provided for @detalleTitle.
  ///
  /// In es, this message translates to:
  /// **'Cliente'**
  String get detalleTitle;

  /// No description provided for @detalleSinTiquetera.
  ///
  /// In es, this message translates to:
  /// **'Sin tiquetera activa.'**
  String get detalleSinTiquetera;

  /// No description provided for @detalleRecargar.
  ///
  /// In es, this message translates to:
  /// **'Recargar'**
  String get detalleRecargar;

  /// No description provided for @detalleRegenerar.
  ///
  /// In es, this message translates to:
  /// **'Regenerar QR'**
  String get detalleRegenerar;

  /// No description provided for @recargaTitle.
  ///
  /// In es, this message translates to:
  /// **'Nueva recarga'**
  String get recargaTitle;

  /// No description provided for @recargaHint.
  ///
  /// In es, this message translates to:
  /// **'Se suma al saldo actual.'**
  String get recargaHint;

  /// No description provided for @recargaOk.
  ///
  /// In es, this message translates to:
  /// **'Recarga registrada.'**
  String get recargaOk;

  /// No description provided for @cancel.
  ///
  /// In es, this message translates to:
  /// **'Cancelar'**
  String get cancel;

  /// No description provided for @regenerarTitle.
  ///
  /// In es, this message translates to:
  /// **'¿Regenerar el código QR?'**
  String get regenerarTitle;

  /// No description provided for @regenerarBody.
  ///
  /// In es, this message translates to:
  /// **'El código QR anterior dejará de funcionar. El cliente necesitará el nuevo.'**
  String get regenerarBody;

  /// No description provided for @regenerarConfirm.
  ///
  /// In es, this message translates to:
  /// **'Regenerar'**
  String get regenerarConfirm;
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  Future<AppLocalizations> load(Locale locale) {
    return SynchronousFuture<AppLocalizations>(lookupAppLocalizations(locale));
  }

  @override
  bool isSupported(Locale locale) =>
      <String>['es'].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'es':
      return AppLocalizationsEs();
  }

  throw FlutterError(
    'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
    'an issue with the localizations generation tool. Please file an issue '
    'on GitHub with a reproducible sample app and the gen-l10n configuration '
    'that was used.',
  );
}
