/// Reglas de entrada de clientes (espejo de `docs/API.md#clientes-admin`;
/// el servidor valida de nuevo).
abstract final class ClienteRules {
  static const int minAlmuerzos = 1;
  static const int maxAlmuerzos = 500;
  static const int nombreMin = 2;
  static const int nombreMax = 120;
  static const int documentoMax = 30;
  static const int emailMax = 254;
  static const String countryPrefix = '+57';
  static const Duration _bogotaOffset = Duration(hours: -5);

  static final RegExp _localPhone = RegExp(r'^\d{10}$');
  static final RegExp _email = RegExp(r'^[^@\s]+@[^@\s]+\.[^@\s]+$');

  static bool validNombre(String v) {
    final t = v.trim();
    return t.length >= nombreMin && t.length <= nombreMax;
  }

  /// Teléfono local colombiano de 10 dígitos (se ignoran espacios y guiones).
  static String digits(String v) => v.replaceAll(RegExp(r'[\s-]'), '');

  static bool validTelefono(String v) => _localPhone.hasMatch(digits(v));

  static String toE164(String local) => '$countryPrefix${digits(local)}';

  static bool validEmail(String v) {
    final t = v.trim();
    return t.isEmpty || (t.length <= emailMax && _email.hasMatch(t));
  }

  static bool validDocumento(String v) => v.trim().length <= documentoMax;

  static int clampAlmuerzos(int v) => v.clamp(minAlmuerzos, maxAlmuerzos);

  /// Fecha de hoy en Bogotá (UTC-5, sin horario de verano).
  static DateTime todayBogota(DateTime now) {
    final b = now.toUtc().add(_bogotaOffset);
    return DateTime(b.year, b.month, b.day);
  }

  /// `+573001234567` → `300 123 4567`.
  static String displayTelefono(String e164) {
    if (!e164.startsWith(countryPrefix) || e164.length != 13) return e164;
    final n = e164.substring(3);
    return '${n.substring(0, 3)} ${n.substring(3, 6)} ${n.substring(6)}';
  }

  static String displayDate(DateTime d) =>
      '${d.day.toString().padLeft(2, '0')}/'
      '${d.month.toString().padLeft(2, '0')}/${d.year}';
}
