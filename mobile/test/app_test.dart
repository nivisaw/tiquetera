import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:tiquetera/app.dart';
import 'package:tiquetera/features/auth/data/auth_repository.dart';
import 'package:tiquetera/features/auth/domain/user.dart';
import 'package:tiquetera/features/auth/presentation/auth_controller.dart';

class _MockRepo extends Mock implements AuthRepository;

void main() {
  testWidgets('TestTiqueteraApp_StoredCajeroSession_OpensEscaner', (
    tester,
  ) async {
    final repo = _MockRepo();
    when(repo.restore).thenAnswer(
      (_) async => const User(id: '2', nombre: 'Caro', rol: Role.cajero),
    );

    await tester.pumpWidget(
      ProviderScope(
        overrides: [authRepositoryProvider.overrideWithValue(repo)],
        child: const TiqueteraApp(),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Escáner'), findsOneWidget);
    expect(find.byKey(const Key('login_submit')), findsNothing);
  });
}
