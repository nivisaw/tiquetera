import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:mocktail/mocktail.dart';
import 'package:tiquetera/core/providers.dart';
import 'package:tiquetera/core/theme/app_theme.dart';
import 'package:tiquetera/features/clientes/data/clientes_repository.dart';
import 'package:tiquetera/features/clientes/domain/cliente.dart';
import 'package:tiquetera/features/clientes/presentation/cliente_detalle_page.dart';
import 'package:tiquetera/features/clientes/presentation/cliente_form_page.dart';
import 'package:tiquetera/features/clientes/presentation/clientes_page.dart';
import 'package:tiquetera/features/clientes/presentation/clientes_providers.dart';
import 'package:tiquetera/features/clientes/presentation/qr_page.dart';
import 'package:tiquetera/l10n/gen/app_localizations.dart';

class MockClientesRepo extends Mock implements ClientesRepository;

/// 2026-11-10 10:00 en Bogotá.
final DateTime fixedNow = DateTime.utc(2026, 11, 10, 15);

const Map<String, Object?> tiqueteraJson = {
  'id': 't1',
  'cliente_id': 'c1',
  'almuerzos_comprados': 20,
  'saldo': 20,
  'fecha_vencimiento': '2026-12-10',
  'activa': true,
  'created_at': '2026-11-10T15:00:00Z',
};

const Map<String, Object?> clienteJson = {
  'id': 'c1',
  'nombre': 'Ana Pérez',
  'telefono': '+573001234567',
  'documento': null,
  'email': null,
  'activo': true,
  'qr_version': 1,
  'created_at': '2026-11-10T15:00:00Z',
  'updated_at': '2026-11-10T15:00:00Z',
};

final Cliente cliente = Cliente.fromJson({
  ...clienteJson,
  'tiquetera': tiqueteraJson,
});

/// Monta las rutas de clientes reales sobre `initial`.
Widget harness(String initial, MockClientesRepo repo, {ThemeData? theme}) {
  final router = GoRouter(
    initialLocation: initial,
    routes: [
      GoRoute(
        path: '/clientes',
        builder: (_, _) => const ClientesPage(),
        routes: [
          GoRoute(path: 'nuevo', builder: (_, _) => const ClienteFormPage()),
          GoRoute(
            path: 'qr',
            builder: (_, s) => QrPage(args: s.extra! as QrArgs),
          ),
          GoRoute(
            path: ':id',
            builder: (_, s) => ClienteDetallePage(id: s.pathParameters['id']!),
          ),
        ],
      ),
    ],
  );
  return ProviderScope(
    overrides: [
      clientesRepositoryProvider.overrideWithValue(repo),
      clockProvider.overrideWithValue(() => fixedNow),
    ],
    child: MaterialApp.router(
      debugShowCheckedModeBanner: false,
      theme: theme ?? AppTheme.light(),
      locale: const Locale('es', 'CO'),
      supportedLocales: AppLocalizations.supportedLocales,
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      routerConfig: router,
    ),
  );
}

/// Pantalla QR aislada (golden).
Widget qrHarness(QrArgs args, {ThemeData? theme}) => MaterialApp(
  debugShowCheckedModeBanner: false,
  theme: theme ?? AppTheme.light(),
  locale: const Locale('es', 'CO'),
  supportedLocales: AppLocalizations.supportedLocales,
  localizationsDelegates: AppLocalizations.localizationsDelegates,
  home: QrPage(args: args),
);
