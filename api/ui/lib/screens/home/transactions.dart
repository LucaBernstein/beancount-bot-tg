import 'package:flutter/material.dart';

import '../../service/api.dart';

class TransactionsWidget extends StatefulWidget {
  final ClientAuthentication authentication;

  const TransactionsWidget({super.key, required this.authentication});

  @override
  State<TransactionsWidget> createState() => _TransactionsWidgetState();
}

class _TransactionsWidgetState extends State<TransactionsWidget> {
  @override
  Widget build(BuildContext context) {
    return Center(
      child: Text('tx go here...'),
    );
  }
}
