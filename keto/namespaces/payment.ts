import { Namespace, Context } from "@ory/keto-namespace-types"

class profile implements Namespace {}

class payment_tenant implements Namespace {
  related: {
    owner: profile[]
    admin: profile[]
    operator: profile[]
    viewer: profile[]

    // Direct permission grants
    send_payment: profile[]
    receive_payment: profile[]
    search_payments: profile[]
    view_payment_status: profile[]
    update_payment_status: profile[]
    release_payment: profile[]
    initiate_prompt: profile[]
    create_payment_link: profile[]
    reconcile: profile[]
  }

  permits = {
    send_payment: (ctx: Context): boolean =>
      this.related.owner.includes(ctx.subject) ||
      this.related.admin.includes(ctx.subject) ||
      this.related.operator.includes(ctx.subject) ||
      this.related.send_payment.includes(ctx.subject),

    receive_payment: (ctx: Context): boolean =>
      this.related.owner.includes(ctx.subject) ||
      this.related.admin.includes(ctx.subject) ||
      this.related.operator.includes(ctx.subject) ||
      this.related.receive_payment.includes(ctx.subject),

    search_payments: (ctx: Context): boolean =>
      this.related.owner.includes(ctx.subject) ||
      this.related.admin.includes(ctx.subject) ||
      this.related.operator.includes(ctx.subject) ||
      this.related.viewer.includes(ctx.subject) ||
      this.related.search_payments.includes(ctx.subject),

    view_payment_status: (ctx: Context): boolean =>
      this.related.owner.includes(ctx.subject) ||
      this.related.admin.includes(ctx.subject) ||
      this.related.operator.includes(ctx.subject) ||
      this.related.viewer.includes(ctx.subject) ||
      this.related.view_payment_status.includes(ctx.subject),

    update_payment_status: (ctx: Context): boolean =>
      this.related.owner.includes(ctx.subject) ||
      this.related.admin.includes(ctx.subject) ||
      this.related.update_payment_status.includes(ctx.subject),

    release_payment: (ctx: Context): boolean =>
      this.related.owner.includes(ctx.subject) ||
      this.related.admin.includes(ctx.subject) ||
      this.related.operator.includes(ctx.subject) ||
      this.related.release_payment.includes(ctx.subject),

    initiate_prompt: (ctx: Context): boolean =>
      this.related.owner.includes(ctx.subject) ||
      this.related.admin.includes(ctx.subject) ||
      this.related.operator.includes(ctx.subject) ||
      this.related.initiate_prompt.includes(ctx.subject),

    create_payment_link: (ctx: Context): boolean =>
      this.related.owner.includes(ctx.subject) ||
      this.related.admin.includes(ctx.subject) ||
      this.related.operator.includes(ctx.subject) ||
      this.related.create_payment_link.includes(ctx.subject),

    reconcile: (ctx: Context): boolean =>
      this.related.owner.includes(ctx.subject) ||
      this.related.admin.includes(ctx.subject) ||
      this.related.reconcile.includes(ctx.subject),
  }
}
