/// Construye enlaces `wa.me` para abrir el chat de WhatsApp de un cliente.
abstract final class WhatsAppLink {
  /// [telefonoE164] admite `+573001234567`; wa.me exige solo dígitos.
  /// Devuelve `null` si el teléfono no tiene dígitos.
  static Uri? build(String telefonoE164, String mensaje) {
    final digits = telefonoE164.replaceAll(RegExp(r'\D'), '');
    if (digits.isEmpty) return null;
    final query = mensaje.isEmpty
        ? ''
        : '?text=${Uri.encodeComponent(mensaje)}';
    return Uri.parse('https://wa.me/$digits$query');
  }
}
