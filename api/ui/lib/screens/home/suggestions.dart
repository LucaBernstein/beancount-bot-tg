import 'package:flutter/material.dart';

import '../../service/api.dart';

class SuggestionsWidget extends StatefulWidget {
  final ClientAuthentication authentication;

  const SuggestionsWidget({super.key, required this.authentication});

  @override
  State<SuggestionsWidget> createState() => _SuggestionsWidgetState();
}

class _SuggestionsWidgetState extends State<SuggestionsWidget> {
  @override
  Widget build(BuildContext context) {
    return const Center(
      // Center is a layout widget. It takes a single child and positions it
      // in the middle of the parent.
      child: SizedBox(
        width: 350,
        child: null,
      ),
    );
  }
}
