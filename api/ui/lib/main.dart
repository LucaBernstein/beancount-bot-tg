import 'package:flutter/material.dart';
import 'package:ui/screens/home.dart';
import 'package:ui/screens/login.dart';

import 'landing.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Beancount-Bot-Tg Web-UI',
      routes: {
        '/': (context) => Landing(),
        '/login': (context) => const Login(),
        '/home': (context) => const Home(),
      },
      theme: ThemeData(
        primarySwatch: Colors.deepOrange,
      ),
    );
  }
}
