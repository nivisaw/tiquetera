import 'package:flutter_test/flutter_test.dart';
import 'package:tiquetera/features/clientes/domain/whatsapp_link.dart';

void main() {
  test('TestWhatsAppLink_E164WithText_DigitsAndEncodedText', () {
    final uri = WhatsAppLink.build('+573001234567', 'Hola Ana & José: 20/12?');
    expect(
      uri.toString(),
      'https://wa.me/573001234567?text=Hola%20Ana%20%26%20Jos%C3%A9%3A%2020%2F12%3F',
    );
  });

  test('TestWhatsAppLink_EmptyMessage_NoQuery', () {
    expect(
      WhatsAppLink.build('+57 300 123 4567', '').toString(),
      'https://wa.me/573001234567',
    );
  });

  test('TestWhatsAppLink_NewlineAndPlus_Encoded', () {
    final uri = WhatsAppLink.build('+573001234567', 'a+b\nc')!;
    expect(uri.query, 'text=a%2Bb%0Ac');
    expect(uri.queryParameters['text'], 'a+b\nc');
  });

  test('TestWhatsAppLink_NoDigits_ReturnsNull', () {
    expect(WhatsAppLink.build('+', 'hola'), isNull);
  });
}
