import 'package:flutter_test/flutter_test.dart';
import 'package:tiquetera/core/router/app_router.dart';
import 'package:tiquetera/features/auth/domain/user.dart';

const _admin = User(id: '1', nombre: 'Ana', rol: Role.admin);
const _cajero = User(id: '2', nombre: 'Caro', rol: Role.cajero);

void main() {
  String? go(String location, {User? user, bool loading = false}) =>
      authRedirect(loading: loading, user: user, location: location);

  test('TestAuthRedirect_Loading_GoesToSplash', () {
    expect(go(AppRoutes.clientes, loading: true), AppRoutes.splash);
    expect(go(AppRoutes.splash, loading: true), isNull);
  });

  test('TestAuthRedirect_NoSession_GoesToLogin', () {
    expect(go(AppRoutes.clientes), AppRoutes.login);
    expect(go(AppRoutes.escaner), AppRoutes.login);
    expect(go(AppRoutes.login), isNull);
  });

  test('TestAuthRedirect_UnknownRole_GoesToLogin', () {
    expect(
      go(
        AppRoutes.escaner,
        user: const User(id: '3', nombre: 'X'),
      ),
      AppRoutes.login,
    );
  });

  test('TestAuthRedirect_Admin_HomeClientesAndNoEscaner', () {
    expect(go(AppRoutes.login, user: _admin), AppRoutes.clientes);
    expect(go(AppRoutes.splash, user: _admin), AppRoutes.clientes);
    expect(go(AppRoutes.escaner, user: _admin), AppRoutes.clientes);
    expect(go(AppRoutes.clientes, user: _admin), isNull);
    expect(go('${AppRoutes.clientes}/abc', user: _admin), isNull);
  });

  test('TestAuthRedirect_Cajero_HomeEscanerAndNoClientes', () {
    expect(go(AppRoutes.login, user: _cajero), AppRoutes.escaner);
    expect(go(AppRoutes.clientes, user: _cajero), AppRoutes.escaner);
    expect(go('${AppRoutes.clientes}/abc', user: _cajero), AppRoutes.escaner);
    expect(go(AppRoutes.escaner, user: _cajero), isNull);
  });
}
