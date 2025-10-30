# Twitch Crypto Donations API

This repository contains the backend API for a Twitch donation service that leverages the Solana blockchain for payments. It provides a secure and efficient way for streamers to receive cryptocurrency donations, with OBS integration for real-time, on-stream alerts.

## Motivation

We started building KapachiPay for a simple reason: we looked for a high-quality, crypto-native donation solution for streamers and couldn't find one. The existing platforms are built for traditional finance, burdened with high fees, slow settlement, and a disconnect from the on-chain world where many creators and their communities now live. Our goal is to provide a seamless, secure, and scalable donation platform built on Solana that empowers creators.

## Team

Our team is composed of top-tier engineers and Solana Day Kazakhstan winners, with a unique blend of technical expertise and real-world experience in the creator economy.

- **Arsen Abdigali**: Backend and Blockchain ([GitHub](https://github.com/abdigaliarsen))
- **Ossein Gizatullayev**: Software Engineer ([GitHub](https://github.com/useing123)) - As a streamer himself, Ossein has an intimate understanding of the creator economy, which drives our user-centric design. ([Link](twitch.tv/arkalis322))
- **Bakdaulet Zharylkassyn**: Software Engineer ([GitHub](https://github.com/bahhhha))
- **Aibar Berekeyev**: Data Science ([GitHub](https://github.com/atropass))

## Resources

- **Website**: [kapachipay.xyz](https://kapachipay.xyz/)
- **Presentation**: [View on Canva](https://www.canva.com/design/DAG3MBP_deQ/ZgOTtirHysh9n3x4ZVMAIQ)
- **Technical Demo**: [Watch on Loom](https://www.loom.com/share/36ccaa04cf5c4c479051cfba6a6d1a0a)
- **X (Twitter)**: [@kapachipay](https://x.com/kapachipay)

## Summary of Features

- **Solana-Based Donations**: Accept donations directly to a Solana wallet, providing a fast and low-cost payment solution.
- **Seamless OBS Integration**: Generate unique widget URLs for OBS to display custom on-stream alerts. We prioritized this for an awesome UX, providing the instant, exciting feedback that makes donation platforms engaging.
- **Secure Wallet Authentication**: Authenticate users by verifying wallet signatures, eliminating the need for traditional passwords.
- **On-Chain Payment Confirmation**: Verify transactions directly on the Solana blockchain to ensure donations are successfully received.
- **User Profile Management**: Allows streamers to manage their public profiles, including username, display name, and avatar.
- **Donation History**: Provides an endpoint for authenticated users to view their donation history.

## Tech Stack & Technical Decisions

- **Language**: Go
- **Framework**: Gin Gonic
- **Blockchain**: Solana (using `gagliardetto/solana-go`)
- **Database**: PostgreSQL (using `lib/pq`)
- **API Documentation**: Swagger (OpenAPI)
- **Containerization**: Docker

We chose **Go** for our backend because it's fast, scalable, and ready for the high-throughput demands of the creator economy. Our **Solana integration** is deep: we don't just send transactions; we use on-chain data to confirm payments and wallet signatures for secure, passwordless authentication. This ensures that the entire process is transparent and trustworthy.

## Target Audience & Traction

Our initial target market is the thousands of creators who are already on-chain but lack the tools to monetize their content effectively. We estimate this to be an immediate market of over 10,000 accounts.

We're already getting positive signals from the community. We are leveraging our network to schedule demos with fellow streamers, and the feedback has been incredibly encouraging.

## Roadmap: The Future of KapachiPay

Donations are just the beginning. Our long-term vision is to build a comprehensive financial toolkit for the creator economy. After implementing KYC and auditing our smart contracts, we plan to expand into:

-   **NFTs**: Allowing creators to mint and sell NFTs directly to their audience.
-   **Real-World Assets (RWA)**: Tokenizing real-world assets for creators.
-   **Yield Farming**: Providing opportunities for creators and their communities to earn yield on their crypto assets.

## Quick Start

The application is live and can be tested on our website. To see it in action, please visit the links below:

- **Live Application**: [kapachipay.xyz/auth](https://kapachipay.xyz/auth)
- **Demonstration Video**: [Watch the Demo](https://www.loom.com/share/36ccaa04cf5c4c479051cfba6a6d1a0a)

Follow the instructions in the video to connect your wallet and send a test donation.
