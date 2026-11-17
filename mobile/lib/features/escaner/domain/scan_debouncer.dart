/// Ignora el mismo QR durante [window] tras aceptarlo (la cámara lo lee
/// varias veces por segundo mientras sigue enfocado).
class ScanDebouncer {
  new(this._now, {this.window = const Duration(seconds: 5)});

  final DateTime Function() _now;
  final Duration window;
  final Map<String, DateTime> _seen = {};

  /// `true` si el token debe procesarse; lo registra en ese caso.
  bool accept(String token) {
    final now = _now();
    _seen.removeWhere((_, t) => now.difference(t) >= window);
    if (_seen.containsKey(token)) return false;
    _seen[token] = now;
    return true;
  }

  /// Permite reintentar de inmediato (p. ej. tras un fallo de red).
  void forget(String token) => _seen.remove(token);
}
