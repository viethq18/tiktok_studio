import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';

import 'api/session.dart';
import 'screens/projects.dart';
import 'screens/sign_in.dart';
import 'widgets/common.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(CarouselStudioApp(api: await Api.load()));
}

class CarouselStudioApp extends StatefulWidget {
  const CarouselStudioApp({super.key, required this.api});

  final Api api;

  @override
  State<CarouselStudioApp> createState() => _CarouselStudioAppState();
}

class _CarouselStudioAppState extends State<CarouselStudioApp> {
  late bool _signedIn = widget.api.client.session != null;

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Carousel Studio',
      debugShowCheckedModeBanner: false,
      theme: buildTheme(),
      // Content language is per project; the interface follows the device.
      supportedLocales: const [Locale('vi'), Locale('en')],
      localizationsDelegates: const [
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      home: _signedIn
          ? ProjectsScreen(
              api: widget.api,
              onSignedOut: () => setState(() => _signedIn = false),
            )
          : SignInScreen(
              api: widget.api,
              onSignedIn: () => setState(() => _signedIn = true),
            ),
    );
  }
}
