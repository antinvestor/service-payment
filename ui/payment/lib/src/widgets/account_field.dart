import 'package:antinvestor_ui_core/widgets/form_field_card.dart';
import 'package:flutter/material.dart';

/// A composite form field for entering account details:
/// account number, country code, and account name.
class AccountField extends StatelessWidget {
  const AccountField({
    super.key,
    required this.label,
    this.description,
    this.accountNumberController,
    this.countryCodeController,
    this.nameController,
    this.readOnly = false,
  });

  final String label;
  final String? description;
  final TextEditingController? accountNumberController;
  final TextEditingController? countryCodeController;
  final TextEditingController? nameController;
  final bool readOnly;

  @override
  Widget build(BuildContext context) {
    return FormFieldCard(
      label: label,
      description: description,
      isRequired: true,
      child: Column(
        children: [
          // Account number
          TextFormField(
            controller: accountNumberController,
            readOnly: readOnly,
            decoration: const InputDecoration(
              labelText: 'Account Number',
              prefixIcon: Icon(Icons.account_balance_wallet, size: 20),
            ),
            validator: (value) {
              if (value == null || value.trim().isEmpty) {
                return 'Account number is required';
              }
              return null;
            },
          ),
          const SizedBox(height: 12),

          // Country code and name row
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Country code
              SizedBox(
                width: 120,
                child: TextFormField(
                  controller: countryCodeController,
                  readOnly: readOnly,
                  decoration: const InputDecoration(
                    labelText: 'Country',
                    hintText: 'KE',
                    prefixIcon: Icon(Icons.flag, size: 20),
                  ),
                  textCapitalization: TextCapitalization.characters,
                  maxLength: 3,
                  validator: (value) {
                    if (value == null || value.trim().isEmpty) {
                      return 'Required';
                    }
                    return null;
                  },
                ),
              ),
              const SizedBox(width: 12),

              // Account name
              Expanded(
                child: TextFormField(
                  controller: nameController,
                  readOnly: readOnly,
                  decoration: const InputDecoration(
                    labelText: 'Account Name',
                    prefixIcon: Icon(Icons.person, size: 20),
                  ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
