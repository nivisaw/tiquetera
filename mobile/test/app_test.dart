import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tiquetera/app.dart';

void main() {
  testWidgets('TestTiqueteraApp_Start_ShowsHomeInSpanish', (tester) async {
    await tester.pumpWidget(const ProviderScope(child: TiqueteraApp()));
    await tester.pumpAndSettle();

    expect(find.text('Tiquetera de almuerzos'), findsOneWidget);
  });
}
