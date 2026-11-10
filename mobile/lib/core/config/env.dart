enum Flavor { dev, prod }

class Env {
  const new({required this.apiBaseUrl, required this.flavor});

  factory fromDartDefines() {
    const url = String.fromEnvironment('API_BASE_URL');
    const flavorName = String.fromEnvironment('FLAVOR', defaultValue: 'dev');
    if (url.isEmpty) {
      throw StateError('API_BASE_URL es obligatorio (--dart-define).');
    }
    final flavor = Flavor.values.byName(flavorName);
    if (flavor == Flavor.prod && !url.startsWith('https://')) {
      throw StateError('En prod API_BASE_URL debe usar https.');
    }
    return Env(apiBaseUrl: url, flavor: flavor);
  }

  final String apiBaseUrl;
  final Flavor flavor;
}
