import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
import 'package:tiquetera/core/theme/app_theme.dart';
import 'package:tiquetera/features/auth/presentation/logout_button.dart';
import 'package:tiquetera/features/escaner/data/camera_permission.dart';
import 'package:tiquetera/features/escaner/presentation/escaner_controller.dart';
import 'package:tiquetera/features/escaner/presentation/resultado_view.dart';
import 'package:tiquetera/l10n/gen/app_localizations.dart';

class EscanerPage extends ConsumerStatefulWidget {
  const new({super.key, this.scannerBuilder});

  /// Sustituye la vista de cámara (tests).
  final Widget Function(void Function(String) onScanned)? scannerBuilder;

  @override
  ConsumerState<EscanerPage> createState() => _EscanerPageState();
}

class _EscanerPageState extends ConsumerState<EscanerPage>
    with WidgetsBindingObserver {
  CameraAccess? _access;

  CameraPermission get _permission => ref.read(cameraPermissionProvider);

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    unawaited(_update(_permission.status));
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed && _access != CameraAccess.granted) {
      unawaited(_update(_permission.status));
    }
  }

  Future<void> _update(Future<CameraAccess> Function() fn) async {
    final s = await fn();
    if (mounted) setState(() => _access = s);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final state = ref.watch(escanerControllerProvider);
    final controller = ref.read(escanerControllerProvider.notifier);
    return switch (state) {
      Exito(:final consumo) => ResultadoView(
        consumo: consumo,
        onDone: controller.volver,
      ),
      Rechazo(:final motivo) => ResultadoView(
        motivo: motivo,
        onDone: controller.volver,
      ),
      _ => Scaffold(
        appBar: AppBar(
          title: Text(l10n.escanerTitle),
          actions: const [LogoutButton()],
        ),
        body: switch (_access) {
          null => const Center(child: CircularProgressIndicator()),
          CameraAccess.granted => _ScannerFrame(
            processing: state is Procesando,
            camera: (widget.scannerBuilder ?? _camera)(controller.onScanned),
          ),
          final denied => _PermissionView(
            blocked: denied == CameraAccess.permanentlyDenied,
            onAllow: () => unawaited(_update(_permission.request)),
            onSettings: () => unawaited(_permission.openSettings()),
          ),
        },
      ),
    };
  }

  Widget _camera(void Function(String) onScanned) =>
      _CameraView(onScanned: onScanned);
}

class _PermissionView extends StatelessWidget {
  const new({
    required this.blocked,
    required this.onAllow,
    required this.onSettings,
  });

  final bool blocked;
  final VoidCallback onAllow;
  final VoidCallback onSettings;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Center(
      child: SingleChildScrollView(
        padding: const EdgeInsets.all(AppSpacing.lg),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const ExcludeSemantics(child: Icon(Icons.photo_camera, size: 64)),
            const SizedBox(height: AppSpacing.md),
            Text(
              l10n.escanerPermisoTitle,
              style: Theme.of(context).textTheme.titleLarge,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: AppSpacing.sm),
            Text(
              blocked ? l10n.escanerPermisoDenied : l10n.escanerPermisoBody,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: AppSpacing.lg),
            FilledButton(
              onPressed: blocked ? onSettings : onAllow,
              child: Text(
                blocked
                    ? l10n.escanerPermisoSettings
                    : l10n.escanerPermisoAllow,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// Marco guía, texto y overlay de "procesando" sobre la cámara.
class _ScannerFrame extends StatelessWidget {
  const new({required this.processing, required this.camera});

  final bool processing;
  final Widget camera;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Stack(
      fit: StackFit.expand,
      children: [
        camera,
        Center(
          child: Container(
            width: 260,
            height: 260,
            decoration: BoxDecoration(
              border: Border.all(color: Colors.white, width: 4),
              borderRadius: BorderRadius.circular(AppRadius.md),
            ),
          ),
        ),
        Positioned(
          left: AppSpacing.md,
          right: AppSpacing.md,
          top: AppSpacing.lg,
          child: Text(
            processing ? l10n.escanerProcesando : l10n.escanerGuia,
            textAlign: TextAlign.center,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 18,
              shadows: [Shadow(blurRadius: 4)],
            ),
          ),
        ),
        if (processing)
          const ColoredBox(
            color: Colors.black38,
            child: Center(child: CircularProgressIndicator()),
          ),
      ],
    );
  }
}

class _CameraView extends StatefulWidget {
  const new({required this.onScanned});

  final void Function(String) onScanned;

  @override
  State<_CameraView> createState() => _CameraViewState();
}

class _CameraViewState extends State<_CameraView> {
  final _controller = MobileScannerController(formats: [BarcodeFormat.qrCode]);

  @override
  void dispose() {
    unawaited(_controller.dispose());
    super.dispose();
  }

  void _onDetect(BarcodeCapture capture) {
    for (final b in capture.barcodes) {
      final raw = b.rawValue;
      if (raw != null) return widget.onScanned(raw);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Stack(
      fit: StackFit.expand,
      children: [
        MobileScanner(controller: _controller, onDetect: _onDetect),
        Positioned(
          bottom: AppSpacing.lg,
          left: 0,
          right: 0,
          child: Center(
            child: IconButton.filled(
              iconSize: 32,
              tooltip: l10n.escanerLinterna,
              onPressed: () => unawaited(_controller.toggleTorch()),
              icon: const Icon(Icons.flashlight_on),
            ),
          ),
        ),
      ],
    );
  }
}
