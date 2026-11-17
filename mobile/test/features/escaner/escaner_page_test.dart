import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:tiquetera/core/network/app_failure.dart';
import 'package:tiquetera/core/providers.dart';
import 'package:tiquetera/core/theme/app_theme.dart';
import 'package:tiquetera/features/escaner/data/camera_permission.dart';
import 'package:tiquetera/features/escaner/data/consumos_repository.dart';
import 'package:tiquetera/features/escaner/domain/consumo.dart';
import 'package:tiquetera/features/escaner/presentation/escaner_controller.dart';
import 'package:tiquetera/features/escaner/presentation/escaner_page.dart';
import 'package:tiquetera/l10n/gen/app_localizations.dart';

class _MockRepo extends Mock implements ConsumosRepository;

class _MockPermission extends Mock implements CameraPermission;

final _ok = ConsumoOk(
  nombre: 'Ana P.',
  saldoRestante: 7,
  fechaVencimiento: DateTime(2026, 12, 31),
);

late void Function(String) _scan;
var _keyCount = 0;

Widget _harness(ConsumosRepository repo, CameraPermission perm) =>
    ProviderScope(
      overrides: [
        consumosRepositoryProvider.overrideWithValue(repo),
        cameraPermissionProvider.overrideWithValue(perm),
        idempotencyKeyGeneratorProvider.overrideWithValue(
          () => 'key-${++_keyCount}',
        ),
        clockProvider.overrideWithValue(() => DateTime(2026, 11, 17, 12)),
      ],
      child: MaterialApp(
        theme: AppTheme.light(),
        locale: const Locale('es', 'CO'),
        supportedLocales: AppLocalizations.supportedLocales,
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        home: EscanerPage(
          scannerBuilder: (onScanned) {
            _scan = onScanned;
            return const ColoredBox(color: Colors.black);
          },
        ),
      ),
    );

void main() {
  late _MockRepo repo;
  late _MockPermission perm;

  setUp(() {
    repo = _MockRepo();
    perm = _MockPermission();
    _keyCount = 0;
    when(perm.status).thenAnswer((_) async => CameraAccess.granted);
  });

  void stubRegistrar(Future<ConsumoOk> Function() answer) => when(
    () => repo.registrar(
      qr: any(named: 'qr'),
      idempotencyKey: any(named: 'idempotencyKey'),
    ),
  ).thenAnswer((_) => answer());

  testWidgets('TestEscanerPage_Success_ShowsSaldoAndAutoReturns', (
    tester,
  ) async {
    stubRegistrar(() async => _ok);
    await tester.pumpWidget(_harness(repo, perm));
    await tester.pump();
    _scan('TQ1.abc');
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 400));

    expect(find.text('¡Almuerzo registrado!'), findsOneWidget);
    expect(find.text('Ana P.'), findsOneWidget);
    final saldo = tester.widget<Text>(find.text('7'));
    expect(saldo.style!.fontSize, greaterThanOrEqualTo(64));
    expect(find.text('Vence el 31/12/2026'), findsOneWidget);
    expect(
      tester.widget<Material>(find.byType(Material).first).color,
      AppColors.success,
    );
    verify(() => repo.registrar(qr: 'TQ1.abc', idempotencyKey: 'key-1'))
        .called(1);

    await tester.pump(const Duration(seconds: 3));
    await tester.pump();
    expect(find.text('Apunta al código QR del cliente'), findsOneWidget);
  });

  testWidgets('TestEscanerPage_SameTokenWithin5s_DebitsOnce', (tester) async {
    stubRegistrar(() async => _ok);
    await tester.pumpWidget(_harness(repo, perm));
    await tester.pump();
    _scan('TQ1.abc');
    await tester.pump();
    await tester.tap(find.text('¡Almuerzo registrado!'));
    await tester.pump();
    _scan('TQ1.abc');
    await tester.pump();
    expect(find.text('Apunta al código QR del cliente'), findsOneWidget);
    verify(
      () => repo.registrar(
        qr: any(named: 'qr'),
        idempotencyKey: any(named: 'idempotencyKey'),
      ),
    ).called(1);
  });

  final cases = <String, (AppFailure, String)>{
    'QR_INVALIDO': (
      const ApiFailure(status: 422, code: 'QR_INVALIDO'),
      'Este código no es una tiquetera válida.',
    ),
    'CLIENTE_NO_EXISTE': (
      const ApiFailure(status: 422, code: 'CLIENTE_NO_EXISTE'),
      'Código desconocido o reemplazado. Pide al cliente su QR más reciente.',
    ),
    'CLIENTE_INACTIVO': (
      const ApiFailure(status: 422, code: 'CLIENTE_INACTIVO'),
      'El cliente está inactivo.',
    ),
    'SIN_TIQUETERA': (
      const ApiFailure(status: 422, code: 'SIN_TIQUETERA'),
      'El cliente no tiene una tiquetera activa.',
    ),
    'TIQUETERA_VENCIDA': (
      const ApiFailure(status: 422, code: 'TIQUETERA_VENCIDA'),
      'La tiquetera está vencida.',
    ),
    'SIN_SALDO': (
      const ApiFailure(status: 422, code: 'SIN_SALDO'),
      'La tiquetera no tiene almuerzos disponibles.',
    ),
    'RATE_LIMIT': (
      const ApiFailure(status: 429, code: 'RATE_LIMITED'),
      'Demasiados escaneos. Espera un minuto.',
    ),
    'NETWORK': (
      const NetworkFailure(),
      'Sin conexión con el servidor. Escanea de nuevo.',
    ),
    'TIMEOUT': (
      const TimeoutFailure(),
      'Sin conexión con el servidor. Escanea de nuevo.',
    ),
    'OTHER': (
      const ApiFailure(status: 409, code: 'IDEMPOTENCY_KEY_CONFLICT'),
      'No se pudo registrar. Inténtalo de nuevo.',
    ),
  };

  for (final MapEntry(key: name, value: (failure, message)) in cases.entries) {
    testWidgets('TestEscanerPage_Rejection${name}_ShowsReason', (tester) async {
      stubRegistrar(() async => throw failure);
      await tester.pumpWidget(_harness(repo, perm));
      await tester.pump();
      _scan('TQ1.abc');
      await tester.pump();

      expect(find.text('Almuerzo no registrado'), findsOneWidget);
      expect(find.text(message), findsOneWidget);
      expect(
        tester.widget<Material>(find.byType(Material).first).color,
        AppColors.rejection,
      );
      await tester.pump(const Duration(seconds: 3));
      await tester.pump();
      expect(find.text('Almuerzo no registrado'), findsNothing);
    });
  }

  testWidgets('TestEscanerPage_NetworkFailure_AllowsImmediateRetry', (
    tester,
  ) async {
    var calls = 0;
    stubRegistrar(() async {
      calls++;
      if (calls == 1) throw const NetworkFailure();
      return _ok;
    });
    await tester.pumpWidget(_harness(repo, perm));
    await tester.pump();
    _scan('TQ1.abc');
    await tester.pump();
    await tester.tap(find.text('Almuerzo no registrado'));
    await tester.pump();
    _scan('TQ1.abc');
    await tester.pump();
    expect(find.text('¡Almuerzo registrado!'), findsOneWidget);
    verify(() => repo.registrar(qr: 'TQ1.abc', idempotencyKey: 'key-2'))
        .called(1);
  });

  testWidgets('TestEscanerPage_PermissionDenied_ExplainsAndRequests', (
    tester,
  ) async {
    when(perm.status).thenAnswer((_) async => CameraAccess.denied);
    when(perm.request).thenAnswer((_) async => CameraAccess.granted);
    await tester.pumpWidget(_harness(repo, perm));
    await tester.pump();
    expect(find.text('Necesitamos la cámara'), findsOneWidget);

    await tester.tap(find.text('Permitir cámara'));
    await tester.pump();
    verify(perm.request).called(1);
    expect(find.text('Apunta al código QR del cliente'), findsOneWidget);
  });

  testWidgets('TestEscanerPage_PermissionBlocked_OpensSettings', (
    tester,
  ) async {
    when(perm.status).thenAnswer((_) async => CameraAccess.permanentlyDenied);
    when(perm.openSettings).thenAnswer((_) async => true);
    await tester.pumpWidget(_harness(repo, perm));
    await tester.pump();

    await tester.tap(find.text('Abrir ajustes'));
    verify(perm.openSettings).called(1);
  });
}
