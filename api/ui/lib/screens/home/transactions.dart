import 'package:flutter/material.dart';

import '../../models/transaction.dart';
import '../../service/api.dart';

class TransactionsWidget extends StatefulWidget {
  final ClientAuthentication authentication;

  const TransactionsWidget({super.key, required this.authentication});

  @override
  State<TransactionsWidget> createState() => _TransactionsWidgetState();
}

class _TransactionsWidgetState extends State<TransactionsWidget> {
  late Future<List<Transaction>> _transactions;

  Future<List<Transaction>> _loadTransactions() async {
    List<Transaction>? transactions;
    String? error;
    (transactions, error) = await widget.authentication.getTransactions();
    if (error != null && error.isNotEmpty) {
      _snackbarMessage(error);
    }
    return transactions;
  }

  void _snackbarMessage(String s) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(s)),
    );
  }

  @override
  void initState() {
    super.initState();
    _transactions = _loadTransactions();
  }

  Future<void> _deleteTx(int id) async {
    await widget.authentication.deleteTransaction(id);
    setState(() {
      _transactions = _loadTransactions();
    });
  }

  @override
  Widget build(BuildContext context) {
    return Center(
      child: FutureBuilder<List<Transaction>>(
        future: _transactions,
        builder: (context, snapshot) {
          if (snapshot.hasData) {
            List<Transaction>? tx = snapshot.data;
            if (tx != null) {
              if (tx.length == 0) {
                return const Text(
                    'Currently, there are no transactions to display with the current filter selection.');
              }
              List<Widget> txList = [];
              for (var t in tx) {
                txList.add(TransactionWidget(
                  tx: t,
                  fnRm: _deleteTx,
                ));
              }
              return SelectionArea(
                  child: Column(
                children: txList,
              ));
            }
          }
          return const Scaffold(
              body: Center(child: CircularProgressIndicator()));
        },
      ),
    );
  }
}

class TransactionWidget extends StatelessWidget {
  final Transaction tx;
  final Future<void> Function(int id) fnRm;

  const TransactionWidget({super.key, required this.tx, required this.fnRm});

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        Text(tx.booking),
        TextButton(
            onPressed: () => fnRm(tx.id),
            child:
                const Icon(Icons.delete_forever_outlined, color: Colors.red)),
      ],
    );
  }
}
