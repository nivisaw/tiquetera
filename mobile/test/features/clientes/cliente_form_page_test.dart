import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:tiquetera/core/network/app_failure.dart';
import 'package:tiquetera/features/clientes/domain/cliente.dart';
import 'package:tiquetera/features/clientes/presentation/qr_page.dart';

import 'fixtures.dart';

Future<MockClientesRepo> _pump(WidgetTester tester) async {
  final repo = MockClientesRepo();
  await tester.pumpWidget(harness('/clientes/nuevo', repo));
  await tester.pumpAndSettle();
  return repo;
}

Future<void> _fill(WidgetTester tester) async {
  await tester.enterText(
    find.byKey(const Key('cliente_nombre')),
    ' Ana Pérez ',
  );
  await tester.enterText(
    find.byKey(const Key('cliente_telefono')),
    '300 123 4567',
  );
}

Future<void> _save(WidgetTester tester) async {
  await tester.ensureVisible(find.byKey(const Key('cliente_save')));
  await tester.tap(find.byKey(const Key('cliente_save')));
  await tester.pump();
}

ClienteCreado _creado() => ClienteCreado(
  cliente: cliente,
  tiquetera: cliente.tiquetera!,
  qrPayload: 'TQ1.abc',
);

void main() {
  setUpAll(
    () => registerFallbackValue(
      NuevoCliente(
        nombre: '',
        telefono: '',
        almuerzos: 1,
        fechaVencimiento: DateTime(2026),
      ),
    ),
  );

  testWidgets('TestClienteForm_Defaults_ShowsTodayPlus30AndPrefix', (
    tester,
  ) async {
    await _pump(tester);
    expect(find.text('10/12/2026'), findsOneWidget);
    expect(find.text('+57 '), findsOneWidget);
    expect(find.text('20'), findsOneWidget);
  });

  testWidgets('TestClienteForm_Invalid_ShowsErrorsAndDoesNotSend', (
    tester,
  ) async {
    final repo = await _pump(tester);
    await tester.enterText(find.byKey(const Key('cliente_telefono')), '123');
    await tester.enterText(find.byKey(const Key('cliente_email')), 'ana@');
    await tester.enterText(find.byKey(const Key('cliente_almuerzos')), '501');
    await _save(tester);

    expect(find.text('Escribe entre 2 y 120 caracteres.'), findsOneWidget);
    expect(
      find.text('Escribe 10 dígitos, por ejemplo 300 123 4567.'),
      findsOneWidget,
    );
    expect(find.text('Correo no válido.'), findsOneWidget);
    expect(find.text('Entre 1 y 500 almuerzos.'), findsOneWidget);
    verifyNever(() => repo.create(any()));
  });

  testWidgets('TestClienteForm_Stepper_ClampsAtBounds', (tester) async {
    await _pump(tester);
    await tester.enterText(find.byKey(const Key('cliente_almuerzos')), '1');
    await tester.pump();
    final minus = tester.widget<IconButton>(
      find.byKey(const Key('almuerzos_minus')),
    );
    expect(minus.onPressed, isNull);
    await tester.tap(find.byKey(const Key('almuerzos_plus')));
    await tester.pump();
    expect(find.text('2'), findsOneWidget);
  });

  testWidgets('TestClienteForm_Valid_SendsNormalizedAndOpensQr', (
    tester,
  ) async {
    final repo = await _pump(tester);
    final pending = Completer<ClienteCreado>();
    when(() => repo.create(any())).thenAnswer((_) => pending.future);

    await _fill(tester);
    await _save(tester);

    final saveBtn = tester.widget<FilledButton>(
      find.byKey(const Key('cliente_save')),
    );
    expect(saveBtn.onPressed, isNull);
    final sent =
        verify(() => repo.create(captureAny())).captured.single as NuevoCliente;
    expect(sent.nombre, 'Ana Pérez');
    expect(sent.telefono, '+573001234567');
    expect(sent.documento, isNull);
    expect(sent.email, isNull);
    expect(sent.almuerzos, 20);
    expect(sent.fechaVencimiento, DateTime(2026, 12, 10));

    pending.complete(_creado());
    await tester.pumpAndSettle();
    expect(find.byType(QrPage), findsOneWidget);
    expect(find.text('20 almuerzos disponibles'), findsOneWidget);
  });

  testWidgets('TestClienteForm_Duplicado_ShowsMessage', (tester) async {
    final repo = await _pump(tester);
    when(() => repo.create(any()))
        .thenThrow(const ApiFailure(status: 409, code: 'TELEFONO_DUPLICADO'));
    await _fill(tester);
    await _save(tester);
    await tester.pumpAndSettle();
    expect(find.text('Ya existe un cliente con ese teléfono.'), findsOneWidget);
  });
}
