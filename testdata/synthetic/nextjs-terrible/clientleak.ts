"use client"

import { useState } from 'react';

export default function LeakyComponent() {
  const secret = process.env.STRIPE_SECRET_KEY;
  return <div>{secret}</div>;
}
