import { Namespace, Context } from "@ory/keto-namespace-types"

class profile implements Namespace {}

class ledger_tenant implements Namespace {
  related: {
    owner: profile[]
    admin: profile[]
    operator: profile[]
    viewer: profile[]

    manage_ledger: profile[]
    view_ledger: profile[]
    manage_account: profile[]
    view_account: profile[]
    create_transaction: profile[]
    reverse_transaction: profile[]
    update_transaction: profile[]
    view_transaction: profile[]
  }

  permits = {
    manage_ledger: (ctx: Context): boolean =>
      this.related.owner.includes(ctx.subject) ||
      this.related.admin.includes(ctx.subject) ||
      this.related.manage_ledger.includes(ctx.subject),

    view_ledger: (ctx: Context): boolean =>
      this.permits.manage_ledger(ctx) ||
      this.related.operator.includes(ctx.subject) ||
      this.related.viewer.includes(ctx.subject) ||
      this.related.view_ledger.includes(ctx.subject),

    manage_account: (ctx: Context): boolean =>
      this.related.owner.includes(ctx.subject) ||
      this.related.admin.includes(ctx.subject) ||
      this.related.operator.includes(ctx.subject) ||
      this.related.manage_account.includes(ctx.subject),

    view_account: (ctx: Context): boolean =>
      this.permits.manage_account(ctx) ||
      this.related.viewer.includes(ctx.subject) ||
      this.related.view_account.includes(ctx.subject),

    create_transaction: (ctx: Context): boolean =>
      this.related.owner.includes(ctx.subject) ||
      this.related.admin.includes(ctx.subject) ||
      this.related.operator.includes(ctx.subject) ||
      this.related.create_transaction.includes(ctx.subject),

    reverse_transaction: (ctx: Context): boolean =>
      this.related.owner.includes(ctx.subject) ||
      this.related.admin.includes(ctx.subject) ||
      this.related.reverse_transaction.includes(ctx.subject),

    update_transaction: (ctx: Context): boolean =>
      this.related.owner.includes(ctx.subject) ||
      this.related.admin.includes(ctx.subject) ||
      this.related.update_transaction.includes(ctx.subject),

    view_transaction: (ctx: Context): boolean =>
      this.related.owner.includes(ctx.subject) ||
      this.related.admin.includes(ctx.subject) ||
      this.related.operator.includes(ctx.subject) ||
      this.related.viewer.includes(ctx.subject) ||
      this.related.view_transaction.includes(ctx.subject),
  }
}
