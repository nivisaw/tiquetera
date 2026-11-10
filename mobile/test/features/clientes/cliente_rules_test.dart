import 'package:flutter_test/flutter_test.dart';
import 'package:tiquetera/features/clientes/domain/cliente.dart';
import 'package:tiquetera/features/clientes/domain/cliente_rules.dart';

void main() {
  test('TestClienteRules_Telefono_ValidatesAndNormalizes', () {
    expect(ClienteRules.validTelefono('300 123 4567'), isTrue);
    expect(ClienteRules.validTelefono('300-123-4567'), isTrue);
    expect(ClienteRules.validTelefono('30012345'), isFalse);
    expect(ClienteRules.validTelefono('30012345678'), isFalse);
    expect(ClienteRules.validTelefono('300abc4567'), isFalse);
    expect(ClienteRules.toE164('300 123 4567'), '+573001234567');
    expect(ClienteRules.displayTelefono('+573001234567'), '300 123 4567');
    expect(ClienteRules.displayTelefono('+14155550100'), '+14155550100');
  });

  test('TestClienteRules_Campos_Validates', () {
    expect(ClienteRules.validNombre(' A '), isFalse);
    expect(ClienteRules.validNombre('Ana'), isTrue);
    expect(ClienteRules.validNombre('a' * 121), isFalse);
    expect(ClienteRules.validEmail(''), isTrue);
    expect(ClienteRules.validEmail('ana@x.co'), isTrue);
    expect(ClienteRules.validEmail('ana@x'), isFalse);
    expect(ClienteRules.validDocumento('1' * 31), isFalse);
    expect(ClienteRules.clampAlmuerzos(0), 1);
    expect(ClienteRules.clampAlmuerzos(501), 500);
  });

  test('TestClienteRules_TodayBogota_UsesUtcMinus5', () {
    expect(
      ClienteRules.todayBogota(DateTime.utc(2026, 11, 11, 4, 59)),
      DateTime(2026, 11, 10),
    );
    expect(
      ClienteRules.todayBogota(DateTime.utc(2026, 11, 11, 5)),
      DateTime(2026, 11, 11),
    );
  });

  test('TestApiDate_RoundTrip_FormatsDdMmYyyy', () {
    final d = parseApiDate('2026-01-05');
    expect(formatApiDate(d), '2026-01-05');
    expect(ClienteRules.displayDate(d), '05/01/2026');
  });

  test('TestNuevoCliente_ToJson_OmitsEmptyOptionals', () {
    final json = NuevoCliente(
      nombre: 'Ana',
      telefono: '+573001234567',
      almuerzos: 20,
      fechaVencimiento: DateTime(2026, 12, 10),
    ).toJson();
    expect(json, {
      'nombre': 'Ana',
      'telefono': '+573001234567',
      'almuerzos': 20,
      'fecha_vencimiento': '2026-12-10',
    });
  });
}
