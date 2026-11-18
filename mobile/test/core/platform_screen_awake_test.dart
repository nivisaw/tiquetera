import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tiquetera/core/platform/screen_awake.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  final messenger =
      TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger;

  tearDown(() => messenger.setMockMethodCallHandler(ScreenAwake.channel, null));

  test('TestScreenAwake_Set_SendsKeepOn', () async {
    final calls = <MethodCall>[];
    messenger.setMockMethodCallHandler(ScreenAwake.channel, (c) async {
      calls.add(c);
      return null;
    });
    await ScreenAwake.set(on: true);
    await ScreenAwake.set(on: false);
    expect(calls.map((c) => '${c.method}:${c.arguments}'), [
      'keepOn:true',
      'keepOn:false',
    ]);
  });

  test('TestScreenAwake_PlatformError_IsIgnored', () async {
    messenger.setMockMethodCallHandler(
      ScreenAwake.channel,
      (_) async => throw PlatformException(code: 'x'),
    );
    await expectLater(ScreenAwake.set(on: true), completes);
  });

  test('TestScreenAwake_NoPlugin_IsIgnored', () async {
    await expectLater(ScreenAwake.set(on: true), completes);
  });
}
