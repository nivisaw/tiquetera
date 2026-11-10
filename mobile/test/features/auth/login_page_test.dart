import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:tiquetera/app.dart';
import 'package:tiquetera/core/network/app_failure.dart';
import 'package:tiquetera/features/auth/data/auth_repository.dart';
import 'package:tiquetera/features/auth/domain/user.dart';
import 'package:tiquetera/features/auth/presentation/auth_controller.dart';

class _MockRepo extends Mock implements AuthRepository;

Future<_MockRepo> _pumpApp(WidgetTester tester) async {
  final repo = _MockRepo();
  when(repo.restore).thenAnswer((_) async => null);
  await tester.pumpWidget(
    ProviderScope(
      overrides: [authRepositoryProvider.overrideWithValue(repo)],
      child: const TiqueteraApp(),
    ),
  );
  await tester.pumpAndSettle();
  return repo;
}

Future<void> _fillAndSubmit(WidgetTester tester) async {
  await tester.enterText(find.byKey(const Key('login_username')), 'ana');
  await tester.enterText(find.byKey(const Key('login_password')), 'clave-123');
  await tester.tap(find.byKey(const Key('login_submit')));
  await tester.pump();
}

void main() {
  testWidgets('TestLoginPage_Empty_ShowsRequired', (tester) async {
    final repo = await _pumpApp(tester);

    await tester.tap(find.byKey(const Key('login_submit')));
    await tester.pump();

    expect(find.text('Campo obligatorio'), findsNWidgets(2));
    verifyNever(
      () => repo.login(
        username: any(named: 'username'),
        password: any(named: 'password'),
      ),
    );
  });

  testWidgets('TestLoginPage_Submitting_DisablesButtonAndShowsLoader', (
    tester,
  ) async {
    final repo = await _pumpApp(tester);
    final pending = Completer<User>();
    when(() => repo.login(username: 'ana', password: 'clave-123'))
        .thenAnswer((_) => pending.future);

    await _fillAndSubmit(tester);

    final button = tester.widget<FilledButton>(
      find.byKey(const Key('login_submit')),
    );
    expect(button.onPressed, isNull);
    expect(find.byType(CircularProgressIndicator), findsOneWidget);
    pending.completeError(const NetworkFailure());
    await tester.pumpAndSettle();
  });

  testWidgets('TestLoginPage_InvalidCredentials_ShowsError', (tester) async {
    final repo = await _pumpApp(tester);
    when(() => repo.login(username: 'ana', password: 'clave-123'))
        .thenThrow(const UnauthorizedFailure(code: 'INVALID_CREDENTIALS'));

    await _fillAndSubmit(tester);
    await tester.pumpAndSettle();

    expect(find.text('Usuario o contraseña incorrectos.'), findsOneWidget);
    expect(find.text('Ingresar'), findsOneWidget);
  });

  testWidgets('TestLoginPage_Offline_ShowsNetworkError', (tester) async {
    final repo = await _pumpApp(tester);
    when(() => repo.login(username: 'ana', password: 'clave-123'))
        .thenThrow(const NetworkFailure());

    await _fillAndSubmit(tester);
    await tester.pumpAndSettle();

    expect(find.textContaining('Sin conexión'), findsOneWidget);
  });

  testWidgets('TestLoginPage_AdminOk_GoesToClientes', (tester) async {
    final repo = await _pumpApp(tester);
    when(() => repo.login(username: 'ana', password: 'clave-123')).thenAnswer(
      (_) async => const User(id: '1', nombre: 'Ana', rol: Role.admin),
    );

    await _fillAndSubmit(tester);
    await tester.pumpAndSettle();

    expect(find.text('Clientes'), findsOneWidget);
  });

  testWidgets('TestLoginPage_CajeroOk_GoesToEscanerAndLogout', (tester) async {
    final repo = await _pumpApp(tester);
    when(() => repo.login(username: 'ana', password: 'clave-123')).thenAnswer(
      (_) async => const User(id: '2', nombre: 'Caro', rol: Role.cajero),
    );
    when(repo.logout).thenAnswer((_) async {});

    await _fillAndSubmit(tester);
    await tester.pumpAndSettle();
    expect(find.text('Escáner'), findsOneWidget);

    await tester.tap(find.byTooltip('Cerrar sesión'));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('login_submit')), findsOneWidget);
  });
}
