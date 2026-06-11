import 'package:antinvestor_api_ledger/antinvestor_api_ledger.dart';
import 'package:flutter/material.dart';

import 'ledger_type_badge.dart';

/// A hierarchical tree view of ledgers showing parent-child relationships.
class LedgerTreeView extends StatelessWidget {
  const LedgerTreeView({super.key, required this.ledgers, this.onLedgerTap});

  final List<Ledger> ledgers;
  final ValueChanged<Ledger>? onLedgerTap;

  @override
  Widget build(BuildContext context) {
    // Build a map of parent -> children
    final childMap = <String, List<Ledger>>{};
    final rootLedgers = <Ledger>[];

    for (final ledger in ledgers) {
      if (ledger.parent.isEmpty) {
        rootLedgers.add(ledger);
      } else {
        childMap.putIfAbsent(ledger.parent, () => []).add(ledger);
      }
    }

    if (rootLedgers.isEmpty && ledgers.isNotEmpty) {
      // If no root ledgers found, show all as flat list
      return ListView.builder(
        shrinkWrap: true,
        physics: const NeverScrollableScrollPhysics(),
        itemCount: ledgers.length,
        itemBuilder: (context, index) => _LedgerTreeNode(
          ledger: ledgers[index],
          childMap: childMap,
          depth: 0,
          onTap: onLedgerTap,
        ),
      );
    }

    return ListView.builder(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      itemCount: rootLedgers.length,
      itemBuilder: (context, index) => _LedgerTreeNode(
        ledger: rootLedgers[index],
        childMap: childMap,
        depth: 0,
        onTap: onLedgerTap,
      ),
    );
  }
}

class _LedgerTreeNode extends StatefulWidget {
  const _LedgerTreeNode({
    required this.ledger,
    required this.childMap,
    required this.depth,
    this.onTap,
  });

  final Ledger ledger;
  final Map<String, List<Ledger>> childMap;
  final int depth;
  final ValueChanged<Ledger>? onTap;

  @override
  State<_LedgerTreeNode> createState() => _LedgerTreeNodeState();
}

class _LedgerTreeNodeState extends State<_LedgerTreeNode> {
  bool _expanded = true;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final children = widget.childMap[widget.ledger.id] ?? [];
    final hasChildren = children.isNotEmpty;
    final indent = widget.depth * 24.0;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        InkWell(
          onTap: () => widget.onTap?.call(widget.ledger),
          borderRadius: BorderRadius.circular(8),
          child: Padding(
            padding: EdgeInsets.only(
              left: 16 + indent,
              right: 16,
              top: 8,
              bottom: 8,
            ),
            child: Row(
              children: [
                // Expand/collapse toggle
                if (hasChildren)
                  GestureDetector(
                    onTap: () => setState(() => _expanded = !_expanded),
                    child: Icon(
                      _expanded ? Icons.expand_more : Icons.chevron_right,
                      size: 20,
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  )
                else
                  const SizedBox(width: 20),
                const SizedBox(width: 8),

                // Ledger type color dot
                Container(
                  width: 8,
                  height: 8,
                  decoration: BoxDecoration(
                    color: ledgerTypeColor(widget.ledger.type),
                    shape: BoxShape.circle,
                  ),
                ),
                const SizedBox(width: 8),

                // Ledger info
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        widget.ledger.id,
                        style: theme.textTheme.bodyMedium?.copyWith(
                          fontWeight: FontWeight.w500,
                        ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                      if (widget.ledger.data.fields.isNotEmpty)
                        Text(
                          _dataPreview(widget.ledger.data.fields),
                          style: theme.textTheme.labelSmall?.copyWith(
                            color: theme.colorScheme.onSurfaceVariant,
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                    ],
                  ),
                ),

                // Type badge
                LedgerTypeBadge(type: widget.ledger.type),

                const SizedBox(width: 4),
                Icon(
                  Icons.chevron_right,
                  size: 18,
                  color: theme.colorScheme.onSurfaceVariant,
                ),
              ],
            ),
          ),
        ),

        // Children
        if (hasChildren && _expanded)
          ...children.map(
            (child) => _LedgerTreeNode(
              ledger: child,
              childMap: widget.childMap,
              depth: widget.depth + 1,
              onTap: widget.onTap,
            ),
          ),

        // Divider
        if (widget.depth == 0)
          Divider(
            height: 1,
            indent: 16,
            endIndent: 16,
            color: theme.colorScheme.outlineVariant,
          ),
      ],
    );
  }

  String _dataPreview(Map<String, dynamic> data) {
    if (data.isEmpty) return '';
    final entries = data.entries.take(3);
    return entries.map((e) => '${e.key}: ${e.value.stringValue}').join(', ');
  }
}
