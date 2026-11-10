import 'package:flutter/foundation.dart';
import 'package:go_router/go_router.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:tiquetera/features/auth/domain/user.dart';
import 'package:tiquetera/features/auth/presentation/auth_controller.dart';
import 'package:tiquetera/features/auth/presentation/login_page.dart';
import 'package:tiquetera/features/auth/presentation/splash_page.dart';
import 'package:tiquetera/features/clientes/presentation/cliente_detalle_page.dart';
import 'package:tiquetera/features/clientes/presentation/cliente_form_page.dart';
import 'package:tiquetera/features/clientes/presentation/clientes_page.dart';
import 'package:tiquetera/features/clientes/presentation/qr_page.dart';
import 'package:tiquetera/features/escaner/presentation/escaner_page.dart';

part 'app_router.g.dart';

abstract final class AppRoutes {
  static const splash = '/';
  static const login = '/login';
  static const clientes = '/clientes';
  static const escaner = '/escaner';
}

const Map<Role, Set<String>> _roleRoutes = {
  Role.admin: {AppRoutes.clientes},
  Role.cajero: {AppRoutes.escaner},
};

String homeFor(Role role) => switch (role) {
  Role.admin => AppRoutes.clientes,
  Role.cajero => AppRoutes.escaner,
};

/// Guard puro: decide a dónde ir según la sesión y la ruta pedida.
/// El rol aquí solo decide la navegación; el servidor valida cada petición.
String? authRedirect({
  required bool loading,
  required User? user,
  required String location,
}) {
  if (loading) return location == AppRoutes.splash ? null : AppRoutes.splash;
  final role = user?.rol;
  if (role == null) return location == AppRoutes.login ? null : AppRoutes.login;
  final allowed = _roleRoutes[role]!;
  final top = '/${Uri.parse(location).pathSegments.firstOrNull ?? ''}';
  return allowed.contains(top) ? null : homeFor(role);
}

@Riverpod(keepAlive: true)
GoRouter appRouter(Ref ref) {
  final refresh = ValueNotifier(0);
  ref
    ..listen(authControllerProvider, (_, _) => refresh.value++)
    ..onDispose(refresh.dispose);
  return GoRouter(
    refreshListenable: refresh,
    redirect: (context, state) {
      final auth = ref.read(authControllerProvider);
      return authRedirect(
        loading: auth.isLoading && !auth.hasValue,
        user: auth.value,
        location: state.matchedLocation,
      );
    },
    routes: [
      GoRoute(path: AppRoutes.splash, builder: (_, _) => const SplashPage()),
      GoRoute(path: AppRoutes.login, builder: (_, _) => const LoginPage()),
      GoRoute(
        path: AppRoutes.clientes,
        builder: (_, _) => const ClientesPage(),
        routes: [
          GoRoute(path: 'nuevo', builder: (_, _) => const ClienteFormPage()),
          GoRoute(
            path: 'qr',
            redirect: (_, state) =>
                state.extra is QrArgs ? null : AppRoutes.clientes,
            builder: (_, state) => QrPage(args: state.extra! as QrArgs),
          ),
          GoRoute(
            path: ':id',
            builder: (_, state) =>
                ClienteDetallePage(id: state.pathParameters['id']!),
          ),
        ],
      ),
      GoRoute(path: AppRoutes.escaner, builder: (_, _) => const EscanerPage()),
    ],
  );
}
